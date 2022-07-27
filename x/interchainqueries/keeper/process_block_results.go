package keeper

import (
	"bytes"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	tendermintLightClientTypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"
	"github.com/neutron-org/neutron/x/interchainqueries/types"
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

// checkHeadersOrder do some basic checks to verify that nextHeader is really next for the header
func checkHeadersOrder(header *tendermintLightClientTypes.Header, nextHeader *tendermintLightClientTypes.Header) error {
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

// VerifyHeaders verify that headers are valid tendermint headers, checks them on validity by trying call ibcClient.UpdateClient(header)
// to update light client's consensus state and checks that they are sequential (tl;dr header.Height + 1 == nextHeader.Height)
func (k Keeper) VerifyHeaders(ctx sdk.Context, clientID string, header exported.Header, nextHeader exported.Header) error {
	// this IBC handler updates the consensus state and the state root from a provided header.
	// But more importantly in the current situation, it checks that header is valid.
	// Honestly we need only to verify headers, but since the check functions are private, and we don't want to duplicate the code,
	// we update consensus state at the same time (because why not?)
	if err := k.ibcKeeper.ClientKeeper.UpdateClient(ctx, clientID, header); err != nil {
		return sdkerrors.Wrapf(err, "failed to update client: %v", err)
	}
	if err := k.ibcKeeper.ClientKeeper.UpdateClient(ctx, clientID, nextHeader); err != nil {
		return sdkerrors.Wrapf(err, "failed to update client: %v", err)
	}

	tmHeader, ok := header.(*tendermintLightClientTypes.Header)
	if !ok {
		return sdkerrors.Wrapf(types.ErrInvalidType, "failed to cast header to tendermint Header")
	}

	tmNextHeader, ok := nextHeader.(*tendermintLightClientTypes.Header)
	if !ok {
		return sdkerrors.Wrapf(types.ErrInvalidType, "failed to cast header to tendermint Header")
	}

	// do some basic check to verify that tmNextHeader is next for the tmHeader
	if err := checkHeadersOrder(tmHeader, tmNextHeader); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidHeader, "block.NextBlockHeader is not next for the block.Header: %v", err)
	}

	return nil
}

// ProcessBlock verifies headers and transaction in the block
func (k Keeper) ProcessBlock(ctx sdk.Context, queryOwner sdk.AccAddress, queryID uint64, clientID string, block *types.Block) error {
	header, err := ibcclienttypes.UnpackHeader(block.Header)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrProtoUnmarshal, "failed to unpack block header: %v", err)
	}

	nextHeader, err := ibcclienttypes.UnpackHeader(block.NextBlockHeader)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrProtoUnmarshal, "failed to unpack block header: %v", err)
	}

	if err := k.VerifyHeaders(ctx, clientID, header, nextHeader); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidHeader, "failed to verify headers: %v", err)
	}

	tmHeader, ok := header.(*tendermintLightClientTypes.Header)
	if !ok {
		ctx.Logger().Debug("ProcessBlock: failed to cast current header to tendermint Header", "query_id", queryID)
		return sdkerrors.Wrap(types.ErrInvalidType, "failed to cast current header to tendermint Header")
	}

	tmNextHeader, ok := nextHeader.(*tendermintLightClientTypes.Header)
	if !ok {
		ctx.Logger().Debug("ProcessBlock: failed to cast next header to tendermint Header", "query_id", queryID)
		return sdkerrors.Wrap(types.ErrInvalidType, "failed to cast next header to tendermint header")
	}

	for _, tx := range block.Txs {
		var txHash = tmtypes.Tx(tx.Data).Hash()
		if !k.CheckTransactionAlreadySubmitted(ctx, queryID, txHash) {
			// Check that cryptography is O.K. (tx is included in the block, tx was executed successfully)
			if err = k.verifyTransaction(tmHeader, tmNextHeader, tx); err != nil {
				ctx.Logger().Debug("ProcessBlock: failed to verifyTransaction",
					"error", err, "query_id", queryID, "tx_hash", hex.EncodeToString(txHash))
				return sdkerrors.Wrapf(types.ErrInternal, "failed to verifyTransaction %s: %v", hex.EncodeToString(txHash), err)
			}

			k.SaveTransactionAsSubmitted(ctx, queryID, txHash)

			// Let the query owner contract process the query result.
			if _, err := k.sudoHandler.SudoTxQueryResult(ctx, queryOwner, queryID, tmHeader.Header.Height, tx.Data); err != nil {
				return sdkerrors.Wrapf(err, "contract %s rejected transaction query result (tx_hash: %s)",
					queryOwner, hex.EncodeToString(txHash))
			}
		} else {
			ctx.Logger().Debug("ProcessBlock: transaction was already submitted",
				"query_id", queryID, "tx_hash", hex.EncodeToString(txHash))
		}
	}

	return nil
}

// verifyTransaction verifies that some transaction is included in block, and the transaction was executed successfully.
// The function checks:
// * transaction is included in block - header.DataHash merkle root contains transactions hash;
// * transactions was executed successfully - transaction's responseDeliveryTx.Code == 0;
// * transaction's responseDeliveryTx is legitimate - nextHeaderLastResultsDataHash merkle root contains
// deterministicResponseDeliverTx(ResponseDeliveryTx).Bytes()
func (k Keeper) verifyTransaction(
	header *tendermintLightClientTypes.Header,
	nextHeader *tendermintLightClientTypes.Header,
	tx *types.TxValue,
) error {
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
