package keeper

import (
	"bytes"
	"encoding/hex"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	tendermintLightClientTypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"
	"github.com/lidofinance/gaia-wasm-zone/x/interchainqueries/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
	tmtypes "github.com/tendermint/tendermint/types"
)

// deterministicResponseDeliverTx strips non-deterministic fields from
// ResponseDeliverTx and returns another ResponseDeliverTx.
func deterministicResponseDeliverTx(response *abci.ResponseDeliverTx) *abci.ResponseDeliverTx {
	return &abci.ResponseDeliverTx{
		Code:      response.Code,
		Data:      response.Data,
		GasWanted: response.GasWanted,
		GasUsed:   response.GasUsed,
	}
}

// checkHeaders do some basic checks to verify that nextHeader is really next for header
func checkHeaders(header *tendermintLightClientTypes.Header, nextHeader *tendermintLightClientTypes.Header) error {
	if nextHeader.Header.Height != header.Header.Height+1 {
		return sdkerrors.Wrapf(types.ErrInvalidHeader, "nextHeader.Height (%d) is not actually next for a header with height %d", nextHeader.Header.Height, header.Header.Height)
	}

	tmHeader, err := tmtypes.HeaderFromProto(header.Header)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidHeader, "failed to get tendermint header from proto header: %v", err)
	}
	tmNextHeader, err := tmtypes.HeaderFromProto(nextHeader.Header)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidHeader, "failed to get tendermint header from proto header: %v", err)
	}

	if !bytes.Equal(tmHeader.NextValidatorsHash, tmNextHeader.ValidatorsHash) {
		return sdkerrors.Wrapf(types.ErrInvalidHeader, "header.NextValidatorsHash is not equal to nextHeader.ValidatorsHash: %s != %s", tmHeader.NextValidatorsHash.String(), tmNextHeader.ValidatorsHash.String())
	}

	if !bytes.Equal(tmHeader.Hash(), tmNextHeader.LastBlockID.Hash) {
		return sdkerrors.Wrapf(types.ErrInvalidHeader, "header.Hash() is not equal to nextHeader.LastBlockID.Hash: %s != %s", tmHeader.Hash().String(), tmNextHeader.LastBlockID.Hash.String())
	}

	return nil
}

// unpackAndVerifyHeaders unpack headers from protobuf, verifies them, updates client's IBC consensus state with them and return the headers on success
func (k Keeper) unpackAndVerifyHeaders(ctx sdk.Context, clientID string, block *types.Block) (*tendermintLightClientTypes.Header, *tendermintLightClientTypes.Header, error) {
	header, err := ibcclienttypes.UnpackHeader(block.Header)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(types.ErrProtoUnmarshal, "failed to unpack block header: %v", err)
	}

	// this IBC handler updates the consensus state and the state root from a provided header.
	// But more importantly in the current situation, it checks that header is valid.
	// Honestly we need only to verify headers, but since the check functions are private, and we don't want to duplicate the code,
	// we update consensus state at the same time (because why not?)
	if err = k.ibcKeeper.ClientKeeper.UpdateClient(ctx, clientID, header); err != nil {
		return nil, nil, sdkerrors.Wrapf(types.ErrInvalidHeader, "failed to vefify header and update client state: %v", err)
	}

	nextHeader, err := ibcclienttypes.UnpackHeader(block.NextBlockHeader)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(types.ErrProtoUnmarshal, "failed to unpack block header: %v", err)
	}

	// this IBC handler updates the consensus state and the state root from a provided header.
	// But more importantly in the current situation, it checks that header is valid.
	// Honestly we need only to verify headers, but since the check functions are private, and we don't want to duplicate the code,
	// we update consensus state at the same time (because why not?)
	if err = k.ibcKeeper.ClientKeeper.UpdateClient(ctx, clientID, nextHeader); err != nil {
		return nil, nil, sdkerrors.Wrapf(types.ErrInvalidHeader, "failed to vefify header and update client state: %v", err)
	}

	tmHeader, ok := header.(*tendermintLightClientTypes.Header)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(types.ErrInvalidType, "failed to cast header to tendermint Header: %v", err)
	}

	tmNextHeader, ok := nextHeader.(*tendermintLightClientTypes.Header)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(types.ErrInvalidType, "failed to cast header to tendermint Header: %v", err)
	}

	// do some basic check to verify that tmNextHeader is next for the tmHeader
	if err = checkHeaders(tmHeader, tmNextHeader); err != nil {
		return nil, nil, sdkerrors.Wrapf(types.ErrInvalidHeader, "block.NextBlockHeader is not next for the block.Header: %v", err)
	}

	return tmHeader, tmNextHeader, nil
}

func verifyTransaction(header *tendermintLightClientTypes.Header, nextHeader *tendermintLightClientTypes.Header, tx *types.TxValue) error {
	// verify inclusion proof
	inclusionProof, err := merkle.ProofFromProto(tx.InclusionProof)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidType, "failed to convert proto proof to merkle proof: %v", err)
	}

	if err = inclusionProof.Verify(header.Header.DataHash, tmtypes.Tx(tx.Data).Hash()); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidProof, "failed to verify inclusion proof: %v", err)
	}

	// verify delivery proof
	deliveryProof, err := merkle.ProofFromProto(tx.DeliveryProof)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidType, "failed to convert proto proof to merkle proof: %v", err)
	}

	responseTx := deterministicResponseDeliverTx(tx.Response)

	responseTxBz, err := responseTx.Marshal()
	if err != nil {
		return sdkerrors.Wrapf(types.ErrProtoMarshal, "failed to marshal ResponseDeliveryTx: %v", err)
	}

	if err = deliveryProof.Verify(nextHeader.Header.LastResultsHash, responseTxBz); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidProof, "failed to verify delivery proof: %v", err)
	}

	// check that transaction was successful
	if tx.Response.Code != abci.CodeTypeOK {
		return sdkerrors.Wrapf(types.ErrInternal, "tx %s is unsuccessful: ResponseDelivery.Code = %d", hex.EncodeToString(tmtypes.Tx(tx.Data).Hash()), tx.Response.Code)
	}

	// check that inclusion proof and delivery proof are for the same transaction
	if deliveryProof.Index != inclusionProof.Index {
		return sdkerrors.Wrapf(types.ErrInvalidProof, "inclusion proof index and delivery proof index are not equal: %d != %d", inclusionProof.Index, deliveryProof.Index)
	}

	return nil
}
