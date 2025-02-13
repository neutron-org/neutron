package keeper

import (
	"bytes"
	"encoding/hex"

	"cosmossdk.io/errors"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	clientkeeper "github.com/cosmos/ibc-go/v8/modules/core/02-client/keeper"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/merkle"
	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	tendermintLightClientTypes "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"

	"github.com/neutron-org/neutron/v5/x/interchainqueries/types"
)

// deterministicExecTxResult strips non-deterministic fields from
// ExecTxResult and returns another ExecTxResult.
func deterministicExecTxResult(response *abci.ExecTxResult) *abci.ExecTxResult {
	return &abci.ExecTxResult{
		Code:      response.Code,
		Data:      response.Data,
		GasWanted: response.GasWanted,
		GasUsed:   response.GasUsed,
	}
}

// checkHeadersOrder do some basic checks to verify that nextHeader is really next for the header
func checkHeadersOrder(header, nextHeader *tendermintLightClientTypes.Header) error {
	if nextHeader.Header.Height != header.Header.Height+1 {
		return errors.Wrapf(types.ErrInvalidHeader, "nextHeader.Height (%d) is not actually next for a header with height %d", nextHeader.Header.Height, header.Header.Height)
	}

	tmHeader, err := tmtypes.HeaderFromProto(header.Header)
	if err != nil {
		return errors.Wrapf(types.ErrInvalidHeader, "failed to get tendermint header from proto header: %v", err)
	}
	tmNextHeader, err := tmtypes.HeaderFromProto(nextHeader.Header)
	if err != nil {
		return errors.Wrapf(types.ErrInvalidHeader, "failed to get tendermint header from proto header: %v", err)
	}

	if !bytes.Equal(tmHeader.NextValidatorsHash, tmNextHeader.ValidatorsHash) {
		return errors.Wrapf(types.ErrInvalidHeader, "header.NextValidatorsHash is not equal to nextHeader.ValidatorsHash: %s != %s", tmHeader.NextValidatorsHash.String(), tmNextHeader.ValidatorsHash.String())
	}

	if !bytes.Equal(tmHeader.Hash(), tmNextHeader.LastBlockID.Hash) {
		return errors.Wrapf(types.ErrInvalidHeader, "header.Hash() is not equal to nextHeader.LastBlockID.Hash: %s != %s", tmHeader.Hash().String(), tmNextHeader.LastBlockID.Hash.String())
	}

	return nil
}

type Verifier struct{}

// VerifyHeaders verify that headers are valid tendermint headers, checks them on validity by trying call ibcClient.UpdateClient(header)
// to update light client's consensus state and checks that they are sequential (tl;dr header.Height + 1 == nextHeader.Height)
func (v Verifier) VerifyHeaders(ctx sdk.Context, clientKeeper clientkeeper.Keeper, clientID string, header, nextHeader exported.ClientMessage) error {
	// this IBC handler updates the consensus state and the state root from a provided header.
	// But more importantly in the current situation, it checks that header is valid.
	// Honestly we need only to verify headers, but since the check functions are private, and we don't want to duplicate the code,
	// we update consensus state at the same time (because why not?)
	if err := clientKeeper.UpdateClient(ctx, clientID, header); err != nil {
		return errors.Wrapf(err, "failed to update client: %v", err)
	}
	if err := clientKeeper.UpdateClient(ctx, clientID, nextHeader); err != nil {
		return errors.Wrapf(err, "failed to update client: %v", err)
	}

	tmHeader, ok := header.(*tendermintLightClientTypes.Header)
	if !ok {
		return errors.Wrapf(types.ErrInvalidType, "failed to cast header to tendermint Header")
	}

	tmNextHeader, ok := nextHeader.(*tendermintLightClientTypes.Header)
	if !ok {
		return errors.Wrapf(types.ErrInvalidType, "failed to cast header to tendermint Header")
	}

	// do some basic check to verify that tmNextHeader is next for the tmHeader
	if err := checkHeadersOrder(tmHeader, tmNextHeader); err != nil {
		return errors.Wrapf(types.ErrInvalidHeader, "block.NextBlockHeader is not next for the block.Header: %v", err)
	}

	return nil
}

func (v Verifier) UnpackHeader(anyHeader *codectypes.Any) (exported.ClientMessage, error) {
	return ibcclienttypes.UnpackClientMessage(anyHeader)
}

// ProcessBlock verifies headers and transaction in the block, and then passes the tx query result to
// the querying contract's sudo handler.
func (k Keeper) ProcessBlock(ctx sdk.Context, queryOwner sdk.AccAddress, queryID uint64, clientID string, block *types.Block) error {
	header, err := k.headerVerifier.UnpackHeader(block.Header)
	if err != nil {
		ctx.Logger().Debug("ProcessBlock: failed to unpack block header", "error", err)
		return errors.Wrapf(types.ErrProtoUnmarshal, "failed to unpack block header: %v", err)
	}

	nextHeader, err := k.headerVerifier.UnpackHeader(block.NextBlockHeader)
	if err != nil {
		ctx.Logger().Debug("ProcessBlock: failed to unpack block header", "error", err)
		return errors.Wrapf(types.ErrProtoUnmarshal, "failed to unpack next block header: %v", err)
	}

	if err := k.headerVerifier.VerifyHeaders(ctx, k.ibcKeeper.ClientKeeper, clientID, header, nextHeader); err != nil {
		ctx.Logger().Debug("ProcessBlock: failed to verify headers", "error", err)
		return errors.Wrapf(types.ErrInvalidHeader, "failed to verify headers: %v", err)
	}

	tmHeader, ok := header.(*tendermintLightClientTypes.Header)
	if !ok {
		ctx.Logger().Debug("ProcessBlock: failed to cast current header to tendermint Header", "query_id", queryID)
		return errors.Wrap(types.ErrInvalidType, "failed to cast current header to tendermint Header")
	}

	tmNextHeader, ok := nextHeader.(*tendermintLightClientTypes.Header)
	if !ok {
		ctx.Logger().Debug("ProcessBlock: failed to cast next header to tendermint Header", "query_id", queryID)
		return errors.Wrap(types.ErrInvalidType, "failed to cast next header to tendermint header")
	}

	var (
		tx     = block.GetTx()
		txData = tx.GetData()
		txHash = tmtypes.Tx(txData).Hash()
	)
	if !k.CheckTransactionIsAlreadyProcessed(ctx, queryID, txHash) {
		// Check that cryptography is O.K. (tx is included in the block, tx was executed successfully)
		if err = k.transactionVerifier.VerifyTransaction(tmHeader, tmNextHeader, tx); err != nil {
			ctx.Logger().Debug("ProcessBlock: failed to verifyTransaction",
				"error", err, "query_id", queryID, "tx_hash", hex.EncodeToString(txHash))
			return errors.Wrapf(types.ErrInternal, "failed to verifyTransaction %s: %v", hex.EncodeToString(txHash), err)
		}

		// Let the query owner contract process the query result.
		if _, err := k.contractManagerKeeper.SudoTxQueryResult(ctx, queryOwner, queryID, ibcclienttypes.NewHeight(tmHeader.TrustedHeight.GetRevisionNumber(), uint64(tmHeader.Header.Height)), txData); err != nil { //nolint:gosec
			ctx.Logger().Debug("ProcessBlock: failed to SudoTxQueryResult",
				"error", err, "query_id", queryID, "tx_hash", hex.EncodeToString(txHash))
			return errors.Wrapf(err, "contract %s rejected transaction query result (tx_hash: %s)",
				queryOwner, hex.EncodeToString(txHash))
		}

		k.SaveTransactionAsProcessed(ctx, queryID, txHash)
	} else {
		ctx.Logger().Debug("ProcessBlock: transaction was already submitted",
			"query_id", queryID, "tx_hash", hex.EncodeToString(txHash))
	}

	return nil
}

type TransactionVerifier struct{}

// VerifyTransaction verifies that some transaction is included in block, and the transaction was executed successfully.
// The function checks:
// * transaction is included in block - header.DataHash merkle root contains transactions hash;
// * transactions was executed successfully - transaction's responseDeliveryTx.Code == 0;
// * transaction's responseDeliveryTx is legitimate - nextHeaderLastResultsDataHash merkle root contains
// deterministicExecTxResult(ResponseDeliveryTx).Bytes()
func (v TransactionVerifier) VerifyTransaction(
	header *tendermintLightClientTypes.Header,
	nextHeader *tendermintLightClientTypes.Header,
	tx *types.TxValue,
) error {
	// verify inclusion proof
	inclusionProof, err := merkle.ProofFromProto(tx.InclusionProof)
	if err != nil {
		return errors.Wrapf(types.ErrInvalidType, "failed to convert proto proof to merkle proof: %v", err)
	}

	if err = inclusionProof.Verify(header.Header.DataHash, tmtypes.Tx(tx.Data).Hash()); err != nil {
		return errors.Wrapf(types.ErrInvalidProof, "failed to verify inclusion proof: %v", err)
	}

	// verify delivery proof
	deliveryProof, err := merkle.ProofFromProto(tx.DeliveryProof)
	if err != nil {
		return errors.Wrapf(types.ErrInvalidType, "failed to convert proto proof to merkle proof: %v", err)
	}

	responseTx := deterministicExecTxResult(tx.Response)

	responseTxBz, err := responseTx.Marshal()
	if err != nil {
		return errors.Wrapf(types.ErrProtoMarshal, "failed to marshal ResponseDeliveryTx: %v", err)
	}

	if err = deliveryProof.Verify(nextHeader.Header.LastResultsHash, responseTxBz); err != nil {
		return errors.Wrapf(types.ErrInvalidProof, "failed to verify delivery proof: %v", err)
	}

	// check that transaction was successful
	if tx.Response.Code != abci.CodeTypeOK {
		return errors.Wrapf(types.ErrInternal, "tx %s is unsuccessful: ResponseDelivery.Code = %d", hex.EncodeToString(tmtypes.Tx(tx.Data).Hash()), tx.Response.Code)
	}

	// check that inclusion proof and delivery proof are for the same transaction
	if deliveryProof.Index != inclusionProof.Index {
		return errors.Wrapf(types.ErrInvalidProof, "inclusion proof index and delivery proof index are not equal: %d != %d", inclusionProof.Index, deliveryProof.Index)
	}

	return nil
}
