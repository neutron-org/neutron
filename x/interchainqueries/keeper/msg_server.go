package keeper

import (
	"context"
	"github.com/cosmos/cosmos-sdk/telemetry"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v3/modules/core/03-connection/types"
	ibccommitmenttypes "github.com/cosmos/ibc-go/v3/modules/core/23-commitment/types"
	"github.com/neutron-org/neutron/x/interchainqueries/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (k msgServer) RegisterInterchainQuery(goCtx context.Context, msg *types.MsgRegisterInterchainQuery) (*types.MsgRegisterInterchainQueryResponse, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), LabelRegisterInterchainQuery)
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.Logger().Debug("Registering interchain query", "msg", msg)

	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Debug("RegisterInterchainQuery: failed to validate message", "message", msg)
		return nil, sdkerrors.Wrapf(err, "invalid msg: %v", err)
	}

	if _, err := k.ibcKeeper.ConnectionKeeper.Connection(goCtx, &ibcconnectiontypes.QueryConnectionRequest{ConnectionId: msg.ConnectionId}); err != nil {
		ctx.Logger().Debug("RegisterInterchainQuery: failed to get connection with ID", "message", msg)
		return nil, sdkerrors.Wrapf(types.ErrInvalidConnectionID, "failed to get connection with ID '%s': %v", msg.ConnectionId, err)
	}

	lastID := k.GetLastRegisteredQueryKey(ctx)
	lastID += 1

	registeredQuery := types.RegisteredQuery{
		Id:                lastID,
		Owner:             msg.Sender,
		QueryData:         msg.QueryData,
		QueryType:         msg.QueryType,
		ZoneId:            msg.ZoneId,
		UpdatePeriod:      msg.UpdatePeriod,
		ConnectionId:      msg.ConnectionId,
		LastEmittedHeight: uint64(ctx.BlockHeight()),
	}

	k.SetLastRegisteredQueryKey(ctx, lastID)
	if err := k.SaveQuery(ctx, registeredQuery); err != nil {
		ctx.Logger().Debug("RegisterInterchainQuery: failed to save query", "message", &msg, "error", err)
		return nil, sdkerrors.Wrapf(err, "failed to save query: %v", err)
	}

	return &types.MsgRegisterInterchainQueryResponse{Id: lastID}, nil
}

func (k msgServer) SubmitQueryResult(goCtx context.Context, msg *types.MsgSubmitQueryResult) (*types.MsgSubmitQueryResultResponse, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), LabelSubmitQueryResult)
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.Logger().Debug("Submitting query result", "message", msg)
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Debug("SubmitQueryResult: invalid msg", err)
		return nil, sdkerrors.Wrapf(err, "invalid msg: %v", err)
	}

	query, err := k.GetQueryByID(ctx, msg.QueryId)
	if err != nil {
		ctx.Logger().Debug("SubmitQueryResult: failed to get query by id", "error", err)
		return nil, sdkerrors.Wrapf(err, "failed to get query by id: %v", err)
	}

	if msg.Result.KvResults != nil {
		resp, err := k.ibcKeeper.ConnectionConsensusState(goCtx, &ibcconnectiontypes.QueryConnectionConsensusStateRequest{
			ConnectionId:   query.ConnectionId,
			RevisionNumber: msg.Result.Revision,
			RevisionHeight: msg.Result.Height + 1,
		})
		if err != nil {
			ctx.Logger().Error("SubmitQueryResult: failed to get consensus state", "error", err)
			return nil, sdkerrors.Wrapf(ibcclienttypes.ErrConsensusStateNotFound, "failed to get consensus state: %v", err)
		}

		consensusState, err := ibcclienttypes.UnpackConsensusState(resp.ConsensusState)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrProtoUnmarshal, "failed to unpack consesus state: %v", err)
		}

		for _, result := range msg.Result.KvResults {
			proof, err := ibccommitmenttypes.ConvertProofs(result.Proof)
			if err != nil {
				ctx.Logger().Debug("SubmitQueryResult: failed to convert crypto.ProofOps to MerkleProof", "error", err)
				return nil, sdkerrors.Wrapf(types.ErrInvalidType, "failed to convert crypto.ProofOps to MerkleProof: %v", err)
			}

			path := ibccommitmenttypes.NewMerklePath(result.StoragePrefix, string(result.Key))

			if err := proof.VerifyMembership(ibccommitmenttypes.GetSDKSpecs(), consensusState.GetRoot(), path, result.Value); err != nil {
				ctx.Logger().Error("SubmitQueryResult: failed to verify proof", "error", err)
				return nil, sdkerrors.Wrapf(types.ErrInvalidProof, "failed to verify proof: %v", err)
			}
		}
	}

	for _, block := range msg.Result.Blocks {
		if err := k.VerifyBlock(ctx, msg.ClientId, block); err != nil {
			ctx.Logger().Debug("SubmitQueryResult: failed to verify block", "error", err)
			return nil, sdkerrors.Wrapf(err, "failed to verify block: %v", err)
		}
	}

	if err = k.SaveQueryResult(ctx, msg.QueryId, msg.Result); err != nil {
		ctx.Logger().Debug("SubmitQueryResult: failed to save query result", "error", err)
		return nil, sdkerrors.Wrapf(err, "failed to save query result: %v", err)
	}

	return &types.MsgSubmitQueryResultResponse{}, nil
}

var _ types.MsgServer = msgServer{}
