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
	ctx.Logger().Debug("RegisterInterchainQuery", "msg", msg)

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
	ctx := sdk.UnwrapSDKContext(goCtx)

	ctx.Logger().Debug("SubmitQueryResult", "message", msg)

	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Debug("SubmitQueryResult: failed to validate message",
			"error", err, "message", msg)
		return nil, sdkerrors.Wrapf(err, "invalid msg: %v", err)
	}

	query, err := k.GetQueryByID(ctx, msg.QueryId)
	if err != nil {
		ctx.Logger().Debug("SubmitQueryResult: failed to GetQueryByID",
			"error", err, "query_id", msg.QueryId)
		return nil, sdkerrors.Wrapf(err, "failed to get query by id: %v", err)
	}

	if msg.Result.KvResults != nil {
		resp, err := k.ibcKeeper.ConnectionConsensusState(goCtx, &ibcconnectiontypes.QueryConnectionConsensusStateRequest{
			ConnectionId:   query.ConnectionId,
			RevisionNumber: msg.Result.Revision,
			RevisionHeight: msg.Result.Height + 1,
		})
		if err != nil {
			ctx.Logger().Debug("SubmitQueryResult: failed to get ConnectionConsensusState",
				"error", err, "query", query, "message", msg)
			return nil, sdkerrors.Wrapf(ibcclienttypes.ErrConsensusStateNotFound, "failed to get consensus state: %v", err)
		}

		consensusState, err := ibcclienttypes.UnpackConsensusState(resp.ConsensusState)
		if err != nil {
			ctx.Logger().Error("SubmitQueryResult: failed to UnpackConsensusState",
				"error", err, "query", query, "message", msg)
			return nil, sdkerrors.Wrapf(types.ErrProtoUnmarshal, "failed to unpack consesus state: %v", err)
		}

		for _, result := range msg.Result.KvResults {
			proof, err := ibccommitmenttypes.ConvertProofs(result.Proof)
			if err != nil {
				ctx.Logger().Debug("SubmitQueryResult: failed to ConvertProofs",
					"error", err, "query", query, "message", msg)
				return nil, sdkerrors.Wrapf(types.ErrInvalidType, "failed to convert crypto.ProofOps to MerkleProof: %v", err)
			}

			path := ibccommitmenttypes.NewMerklePath(result.StoragePrefix, string(result.Key))

			if err := proof.VerifyMembership(ibccommitmenttypes.GetSDKSpecs(), consensusState.GetRoot(), path, result.Value); err != nil {
				ctx.Logger().Debug("SubmitQueryResult: failed to VerifyMembership",
					"error", err, "query", query, "message", msg, "path", path)
				return nil, sdkerrors.Wrapf(types.ErrInvalidProof, "failed to verify proof: %v", err)
			}
		}

		if err = k.SaveKVQueryResult(ctx, msg.QueryId, msg.Result); err != nil {
			ctx.Logger().Error("SubmitQueryResult: failed to SaveKVQueryResult",
				"error", err, "query", query, "message", msg)
			return nil, sdkerrors.Wrapf(err, "failed to SaveKVQueryResult: %v", err)
		}
	}
	if msg.Result.Block != nil && msg.Result.Block.Tx != nil {
		queryOwner, err := sdk.AccAddressFromBech32(query.Owner)
		if err != nil {
			ctx.Logger().Error("SubmitQueryResult: failed to decode AccAddressFromBech32",
				"error", err, "query", query, "message", msg)
			return nil, sdkerrors.Wrapf(types.ErrInternal, "failed to decode owner contract address: %v", err)
		}

		if err := k.ProcessBlock(ctx, queryOwner, msg.QueryId, msg.ClientId, msg.Result.Block); err != nil {
			ctx.Logger().Debug("SubmitQueryResult: failed to ProcessBlock",
				"error", err, "query", query, "message", msg)
			return nil, sdkerrors.Wrapf(err, "failed to ProcessBlock: %v", err)
		}
	}

	return &types.MsgSubmitQueryResultResponse{}, nil
}

var _ types.MsgServer = msgServer{}
