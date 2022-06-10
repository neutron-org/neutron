package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		return fmt.Errorf("nextHeader.Height (%d) is not actually next for a header with height %d", nextHeader.Header.Height, header.Header.Height)
	}

	tmHeader, err := tmtypes.HeaderFromProto(header.Header)
	if err != nil {
		return fmt.Errorf("failed to get tendermint header from proto header: %w", err)
	}
	tmNextHeader, err := tmtypes.HeaderFromProto(nextHeader.Header)
	if err != nil {
		return fmt.Errorf("failed to get tendermint header from proto header: %w", err)
	}

	if !bytes.Equal(tmHeader.NextValidatorsHash, tmNextHeader.ValidatorsHash) {
		return fmt.Errorf("header.NextValidatorsHash is not equal to nextHeader.ValidatorsHash: %s != %s", tmHeader.NextValidatorsHash.String(), tmNextHeader.ValidatorsHash.String())
	}

	if !bytes.Equal(tmHeader.Hash(), tmNextHeader.LastBlockID.Hash) {
		return fmt.Errorf("header.Hash() is not equal to nextHeader.LastBlockID.Hash: %s != %s", tmHeader.Hash().String(), tmNextHeader.LastBlockID.Hash.String())
	}

	return nil
}

// unpackAndVerifyHeaders unpack headers from protobuf, verifies them, updates client's IBC consensus state with them and return the headers on success
func (k Keeper) unpackAndVerifyHeaders(ctx sdk.Context, clientID string, block *types.Block) (*tendermintLightClientTypes.Header, *tendermintLightClientTypes.Header, error) {
	header, err := ibcclienttypes.UnpackHeader(block.Header)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unpack block header: %w", err)
	}

	// this IBC handler updates the consensus state and the state root from a provided header.
	// But more importantly in the current situation, it checks that header is valid.
	// Honestly we need only to verify headers, but since the check functions are private, and we don't want to duplicate the code,
	// we update consensus state at the same time (because why not?)
	if err = k.ibcKeeper.ClientKeeper.UpdateClient(ctx, clientID, header); err != nil {
		return nil, nil, fmt.Errorf("failed to vefify header and update client state: %w", err)
	}

	nextHeader, err := ibcclienttypes.UnpackHeader(block.NextBlockHeader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unpack block header: %w", err)
	}

	// this IBC handler updates the consensus state and the state root from a provided header.
	// But more importantly in the current situation, it checks that header is valid.
	// Honestly we need only to verify headers, but since the check functions are private, and we don't want to duplicate the code,
	// we update consensus state at the same time (because why not?)
	if err = k.ibcKeeper.ClientKeeper.UpdateClient(ctx, clientID, nextHeader); err != nil {
		return nil, nil, fmt.Errorf("failed to vefify header and update client state: %w", err)
	}

	tmHeader, ok := header.(*tendermintLightClientTypes.Header)
	if !ok {
		return nil, nil, fmt.Errorf("failed to cast header to tendermint Header: %w", err)
	}

	tmNextHeader, ok := nextHeader.(*tendermintLightClientTypes.Header)
	if !ok {
		return nil, nil, fmt.Errorf("failed to cast header to tendermint Header: %w", err)
	}

	// do some basic check to verify that tmNextHeader is next for the tmHeader
	if err = checkHeaders(tmHeader, tmNextHeader); err != nil {
		return nil, nil, fmt.Errorf("block.NextBlockHeader is not next for the block.Header: %w", err)
	}

	return tmHeader, tmNextHeader, nil
}

func verifyTransaction(header *tendermintLightClientTypes.Header, nextHeader *tendermintLightClientTypes.Header, tx *types.TxValue) error {
	// verify inclusion proof
	inclusionProof, err := merkle.ProofFromProto(tx.InclusionProof)
	if err != nil {
		return fmt.Errorf("failed to convert proto proof to merkle proof: %w", err)
	}

	if err = inclusionProof.Verify(header.Header.DataHash, tmtypes.Tx(tx.Data).Hash()); err != nil {
		return fmt.Errorf("failed to verify inclusion proof: %w", err)
	}

	// verify delivery proof
	deliveryProof, err := merkle.ProofFromProto(tx.DeliveryProof)
	if err != nil {
		return fmt.Errorf("failed to convert proto proof to merkle proof: %w", err)
	}

	responseTx := deterministicResponseDeliverTx(tx.Response)

	responseTxBz, err := responseTx.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal ResponseDeliveryTx: %w", err)
	}

	if err = deliveryProof.Verify(nextHeader.Header.LastResultsHash, responseTxBz); err != nil {
		return fmt.Errorf("failed to verify delivery proof: %w", err)
	}

	// check that transaction was successful
	if tx.Response.Code != abci.CodeTypeOK {
		return fmt.Errorf("tx %s is unsuccessful: ResponseDelivery.Code = %d", hex.EncodeToString(tmtypes.Tx(tx.Data).Hash()), tx.Response.Code)
	}

	// check that inclusion proof and delivery proof are for the same transaction
	if deliveryProof.Index != inclusionProof.Index {
		return fmt.Errorf("inclusion proof index and delivery proof index are not equal: %d != %d", inclusionProof.Index, deliveryProof.Index)
	}

	return nil
}
