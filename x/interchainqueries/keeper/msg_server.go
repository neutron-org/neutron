package keeper

import (
	"bytes"
	"context"
	"time"

	ics23 "github.com/confio/ics23/go"
	"github.com/cosmos/cosmos-sdk/telemetry"

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
		Id:                 lastID,
		Owner:              msg.Sender,
		TransactionsFilter: msg.TransactionsFilter,
		Keys:               msg.Keys,
		QueryType:          msg.QueryType,
		ZoneId:             msg.ZoneId,
		UpdatePeriod:       msg.UpdatePeriod,
		ConnectionId:       msg.ConnectionId,
		LastEmittedHeight:  uint64(ctx.BlockHeight()),
	}

	k.SetLastRegisteredQueryKey(ctx, lastID)
	if err := k.SaveQuery(ctx, registeredQuery); err != nil {
		ctx.Logger().Debug("RegisterInterchainQuery: failed to save query", "message", &msg, "error", err)
		return nil, sdkerrors.Wrapf(err, "failed to save query: %v", err)
	}

	return &types.MsgRegisterInterchainQueryResponse{Id: lastID}, nil
}

func (k msgServer) SubmitQueryResult(goCtx context.Context, msg *types.MsgSubmitQueryResult) (*types.MsgSubmitQueryResultResponse, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), LabelRegisterInterchainQuery)

	ctx := sdk.UnwrapSDKContext(goCtx)

	ctx.Logger().Debug("SubmitQueryResult", "query_id", msg.QueryId)

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

	if len(msg.Result.KvResults) != len(query.Keys) {
		return nil, sdkerrors.Wrapf(types.ErrInvalidSubmittedResult, "KV keys length from result is not equal to registered query keys length: %v != %v", len(msg.Result.KvResults), query.Keys)
	}

	queryOwner, err := sdk.AccAddressFromBech32(query.Owner)
	if err != nil {
		ctx.Logger().Error("SubmitQueryResult: failed to decode AccAddressFromBech32",
			"error", err, "query", query, "message", msg)
		return nil, sdkerrors.Wrapf(err, "failed to decode owner contract address (%s)", query.Owner)
	}

	if msg.Result.KvResults != nil {
		if !types.InterchainQueryType(query.QueryType).IsKV() {
			return nil, sdkerrors.Wrapf(types.ErrInvalidType, "invalid query result for query type: %s", query.QueryType)
		}

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

		for index, result := range msg.Result.KvResults {
			proof, err := ibccommitmenttypes.ConvertProofs(result.Proof)
			if err != nil {
				ctx.Logger().Debug("SubmitQueryResult: failed to ConvertProofs",
					"error", err, "query", query, "message", msg)
				return nil, sdkerrors.Wrapf(types.ErrInvalidType, "failed to convert crypto.ProofOps to MerkleProof: %v", err)
			}

			if !bytes.Equal(result.Key, query.Keys[index].Key) {
				return nil, sdkerrors.Wrapf(types.ErrInvalidSubmittedResult, "KV key from result is not equal to registered query key: %v != %v", result.Key, query.Keys[index].Key)
			}

			if result.StoragePrefix != query.Keys[index].Path {
				return nil, sdkerrors.Wrapf(types.ErrInvalidSubmittedResult, "KV path from result is not equal to registered query storage prefix: %v != %v", result.StoragePrefix, query.Keys[index].Path)
			}

			path := ibccommitmenttypes.NewMerklePath(result.StoragePrefix, string(result.Key))

			switch proof.GetProofs()[0].GetProof().(type) {
			case *ics23.CommitmentProof_Nonexist:
				if err := proof.VerifyNonMembership(ibccommitmenttypes.GetSDKSpecs(), consensusState.GetRoot(), path); err != nil {
					ctx.Logger().Debug("SubmitQueryResult: failed to VerifyNonMembership",
						"error", err, "query", query, "message", msg, "path", path)
					return nil, sdkerrors.Wrapf(types.ErrInvalidProof, "failed to verify proof: %v", err)
				}
				result.Value = nil
			case *ics23.CommitmentProof_Exist:
				if err := proof.VerifyMembership(ibccommitmenttypes.GetSDKSpecs(), consensusState.GetRoot(), path, result.Value); err != nil {
					ctx.Logger().Debug("SubmitQueryResult: failed to VerifyMembership",
						"error", err, "query", query, "message", msg, "path", path)
					return nil, sdkerrors.Wrapf(types.ErrInvalidProof, "failed to verify proof: %v", err)
				}
			default:
				return nil, sdkerrors.Wrap(types.ErrInvalidProof, "unknown proof type")
			}

			queryOwner, err := sdk.AccAddressFromBech32(query.Owner)
			if err != nil {
				ctx.Logger().Error("SubmitQueryResult: failed to decode AccAddressFromBech32",
					"error", err, "query", query, "message", msg)
				return nil, sdkerrors.Wrapf(types.ErrInternal, "failed to decode owner contract address: %v", err)
			}

			if err = k.SaveKVQueryResult(ctx, msg.QueryId, msg.Result); err != nil {
				ctx.Logger().Error("SubmitQueryResult: failed to SaveKVQueryResult",
					"error", err, "query", query, "message", msg)
				return nil, sdkerrors.Wrapf(err, "failed to SaveKVQueryResult: %v", err)
			}

			if msg.Result.GetAllowKvCallbacks() {
				// Let the query owner contract process the query result.
				if _, err := k.sudoHandler.SudoKVQueryResult(ctx, queryOwner, query.Id); err != nil {
					ctx.Logger().Debug("ProcessBlock: failed to SudoKVQueryResult",
						"error", err, "query_id", query.GetId())
					return nil, sdkerrors.Wrapf(err, "contract %s rejected KV query result (query_id: %d)",
						queryOwner, query.GetId())
				}
				return &types.MsgSubmitQueryResultResponse{}, nil
			}
		}
	}

	if msg.Result.Block != nil && msg.Result.Block.Tx != nil {
		if !types.InterchainQueryType(query.QueryType).IsTX() {
			return nil, sdkerrors.Wrapf(types.ErrInvalidType, "invalid query result for query type: %s", query.QueryType)
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
