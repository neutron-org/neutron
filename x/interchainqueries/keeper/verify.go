package keeper

import (
	"bytes"
	"encoding/hex"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	tendermintLightClientTypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"
	"github.com/lidofinance/gaia-wasm-zone/x/interchainqueries/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/light"
	tmtypes "github.com/tendermint/tendermint/types"
	"math"
	"time"
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

func (k Keeper) VerifyHeaders(ctx sdk.Context, clientID string, header exported.Header, nextHeader exported.Header) error {
	if err := k.checkHeader(ctx, clientID, header); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidHeader, "failed to verify header: %v", err)
	}

	if err := k.checkHeader(ctx, clientID, nextHeader); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidHeader, "failed to verify next header: %v", err)
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

// VerifyBlock verifies headers and transaction in the block
func (k Keeper) VerifyBlock(ctx sdk.Context, clientID string, block *types.Block) error {
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
		return sdkerrors.Wrapf(types.ErrInvalidType, "failed to cast header to tendermint Header: %v", err)
	}

	tmNextHeader, ok := nextHeader.(*tendermintLightClientTypes.Header)
	if !ok {
		return sdkerrors.Wrapf(types.ErrInvalidType, "failed to cast header to tendermint Header: %v", err)
	}

	for _, tx := range block.Txs {
		if err = verifyTransaction(tmHeader, tmNextHeader, tx); err != nil {
			return sdkerrors.Wrapf(types.ErrInternal, "failed to verify transaction %s: %v", hex.EncodeToString(tmtypes.Tx(tx.Data).Hash()), err)
		}
	}

	return nil
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

// checkHeader checks header on validity by trying to find trusted consensus state for it
func (k Keeper) checkHeader(ctx sdk.Context, clientID string, header exported.Header) error {
	tmHeader, ok := header.(*tendermintLightClientTypes.Header)
	if !ok {
		return sdkerrors.Wrapf(types.ErrInvalidType, "failed to cast header to tendermint Header")
	}

	clientStateInterface, found := k.ibcKeeper.ClientKeeper.GetClientState(ctx, clientID)
	if !found {
		return sdkerrors.Wrapf(ibcclienttypes.ErrClientNotFound, "cannot get client with ID %s", clientID)
	}

	clientState, ok := clientStateInterface.(*tendermintLightClientTypes.ClientState)
	if !ok {
		return sdkerrors.Wrapf(types.ErrInvalidType, "cannot cast exported.ClientState interface into ClientState struct")
	}

	clientStore := k.ibcKeeper.ClientKeeper.ClientStore(ctx, clientID)

	trustedConsState, err := tendermintLightClientTypes.GetConsensusState(clientStore, k.cdc, tmHeader.TrustedHeight)
	if err != nil {
		return sdkerrors.Wrapf(
			err, "could not get consensus state from client store at TrustedHeight: %s", tmHeader.TrustedHeight,
		)
	}

	// we can't verify an old header with with some newer consensus state
	// in this case we are trying to find some old consensus state
	if tmHeader.GetHeight().LTE(tmHeader.TrustedHeight) {
		consensusStatesResponse, err := k.ibcKeeper.ClientKeeper.ConsensusStates(sdk.WrapSDKContext(ctx), &ibcclienttypes.QueryConsensusStatesRequest{
			ClientId: clientID,
			Pagination: &query.PageRequest{
				Limit:      math.MaxUint64,
				Reverse:    true,
				CountTotal: true,
			},
		})
		if err != nil {
			return sdkerrors.Wrapf(err, "failed to get consensus states for client with ID: %s", clientID)
		}

		for _, cs := range consensusStatesResponse.GetConsensusStates() {
			if tmHeader.GetHeight().GT(cs.Height) {
				consensusStateInterface, err := ibcclienttypes.UnpackConsensusState(cs.ConsensusState)
				if err != nil {
					return sdkerrors.Wrapf(types.ErrProtoUnmarshal, "failed to unpack consensus state: %v", err)
				}

				consensusState, ok := consensusStateInterface.(*tendermintLightClientTypes.ConsensusState)
				if !ok {
					return sdkerrors.Wrapf(types.ErrInvalidType, "failed to cast exported.ConsensusState to ConsensusState struct")
				}

				tmHeader.TrustedHeight = cs.Height

				return checkValidity(clientState, consensusState, tmHeader, ctx.BlockTime())
			}
		}
		return sdkerrors.Wrapf(types.ErrInvalidHeight, "cannot find satisfying ConsensusState")
	}

	return checkValidity(clientState, trustedConsState, tmHeader, ctx.BlockTime())
}

// checkTrustedHeader checks that consensus state matches trusted fields of Header
// copypasted from ibc-go repo
// https://github.com/cosmos/ibc-go/blob/1be4519a4129d71a462557d01ec4b9e9bd7c196e/modules/light-clients/07-tendermint/types/update.go#L146
func checkTrustedHeader(header *tendermintLightClientTypes.Header, consState *tendermintLightClientTypes.ConsensusState) error {
	tmTrustedValidators, err := tmtypes.ValidatorSetFromProto(header.TrustedValidators)
	if err != nil {
		return sdkerrors.Wrap(err, "trusted validator set in not tendermint validator set type")
	}

	// assert that trustedVals is NextValidators of last trusted header
	// to do this, we check that trustedVals.Hash() == consState.NextValidatorsHash
	tvalHash := tmTrustedValidators.Hash()
	if !bytes.Equal(consState.NextValidatorsHash, tvalHash) {
		return sdkerrors.Wrapf(
			tendermintLightClientTypes.ErrInvalidValidatorSet,
			"trusted validators %s, does not hash to latest trusted validators. Expected: %X, got: %X",
			header.TrustedValidators, consState.NextValidatorsHash, tvalHash,
		)
	}
	return nil
}

// checkValidity checks if the Tendermint header is valid.
// CONTRACT: consState.Height == header.TrustedHeight
// copypasted from ibc-go repo
// https://github.com/cosmos/ibc-go/blob/1be4519a4129d71a462557d01ec4b9e9bd7c196e/modules/light-clients/07-tendermint/types/update.go#L167
func checkValidity(
	clientState *tendermintLightClientTypes.ClientState, consState *tendermintLightClientTypes.ConsensusState,
	header *tendermintLightClientTypes.Header, currentTimestamp time.Time,
) error {
	if err := checkTrustedHeader(header, consState); err != nil {
		return err
	}

	// UpdateClient only accepts updates with a header at the same revision
	// as the trusted consensus state
	if header.GetHeight().GetRevisionNumber() != header.TrustedHeight.RevisionNumber {
		return sdkerrors.Wrapf(
			tendermintLightClientTypes.ErrInvalidHeaderHeight,
			"header height revision %d does not match trusted header revision %d",
			header.GetHeight().GetRevisionNumber(), header.TrustedHeight.RevisionNumber,
		)
	}

	tmTrustedValidators, err := tmtypes.ValidatorSetFromProto(header.TrustedValidators)
	if err != nil {
		return sdkerrors.Wrap(err, "trusted validator set in not tendermint validator set type")
	}

	tmSignedHeader, err := tmtypes.SignedHeaderFromProto(header.SignedHeader)
	if err != nil {
		return sdkerrors.Wrap(err, "signed header in not tendermint signed header type")
	}

	tmValidatorSet, err := tmtypes.ValidatorSetFromProto(header.ValidatorSet)
	if err != nil {
		return sdkerrors.Wrap(err, "validator set in not tendermint validator set type")
	}

	// assert header height is newer than consensus state
	if header.GetHeight().LTE(header.TrustedHeight) {
		return sdkerrors.Wrapf(
			ibcclienttypes.ErrInvalidHeader,
			"header height ≤ consensus state height (%s ≤ %s)", header.GetHeight(), header.TrustedHeight,
		)
	}

	chainID := clientState.GetChainID()
	// If chainID is in revision format, then set revision number of chainID with the revision number
	// of the header we are verifying
	// This is useful if the update is at a previous revision rather than an update to the latest revision
	// of the client.
	// The chainID must be set correctly for the previous revision before attempting verification.
	// Updates for previous revisions are not supported if the chainID is not in revision format.
	if ibcclienttypes.IsRevisionFormat(chainID) {
		chainID, _ = ibcclienttypes.SetRevisionNumber(chainID, header.GetHeight().GetRevisionNumber())
	}

	// Construct a trusted header using the fields in consensus state
	// Only Height, Time, and NextValidatorsHash are necessary for verification
	trustedHeader := tmtypes.Header{
		ChainID:            chainID,
		Height:             int64(header.TrustedHeight.RevisionHeight),
		Time:               consState.Timestamp,
		NextValidatorsHash: consState.NextValidatorsHash,
	}
	signedHeader := tmtypes.SignedHeader{
		Header: &trustedHeader,
	}

	// Verify next header with the passed-in trustedVals
	// - asserts trusting period not passed
	// - assert header timestamp is not past the trusting period
	// - assert header timestamp is past latest stored consensus state timestamp
	// - assert that a TrustLevel proportion of TrustedValidators signed new Commit
	err = light.Verify(
		&signedHeader,
		tmTrustedValidators, tmSignedHeader, tmValidatorSet,
		clientState.TrustingPeriod, currentTimestamp, clientState.MaxClockDrift, clientState.TrustLevel.ToTendermint(),
	)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to verify header")
	}

	return nil
}
