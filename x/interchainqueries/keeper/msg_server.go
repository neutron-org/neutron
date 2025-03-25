package keeper

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"

	"cosmossdk.io/errors"
	tendermint "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	ibccommitmenttypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	ics23 "github.com/cosmos/ics23/go"

	"github.com/neutron-org/neutron/v6/x/interchainqueries/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (m msgServer) RegisterInterchainQuery(goCtx context.Context, msg *types.MsgRegisterInterchainQuery) (*types.MsgRegisterInterchainQueryResponse, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), LabelRegisterInterchainQuery)
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.Logger().Debug("RegisterInterchainQuery", "msg", msg)
	params := m.GetParams(ctx)

	if err := msg.Validate(params); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgRegisterInterchainQuery")
	}

	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		m.Logger(ctx).Debug("RegisterInterchainQuery: failed to parse sender address", "sender_address", msg.Sender)
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse address: %s", msg.Sender)
	}

	if !m.contractManagerKeeper.HasContractInfo(ctx, senderAddr) {
		m.Logger(ctx).Debug("RegisterInterchainQuery: contract not found", "sender_address", msg.Sender)
		return nil, errors.Wrapf(types.ErrNotContract, "%s is not a contract address", msg.Sender)
	}

	if _, err := m.ibcKeeper.ConnectionKeeper.Connection(goCtx, &ibcconnectiontypes.QueryConnectionRequest{ConnectionId: msg.ConnectionId}); err != nil {
		ctx.Logger().Debug("RegisterInterchainQuery: failed to get connection with ID", "message", msg)
		return nil, errors.Wrapf(types.ErrInvalidConnectionID, "failed to get connection with ID '%s': %v", msg.ConnectionId, err)
	}

	lastID := m.GetLastRegisteredQueryKey(ctx)
	lastID++

	registeredQuery := &types.RegisteredQuery{
		Id:                 lastID,
		Owner:              msg.Sender,
		TransactionsFilter: msg.TransactionsFilter,
		Keys:               msg.Keys,
		QueryType:          msg.QueryType,
		UpdatePeriod:       msg.UpdatePeriod,
		ConnectionId:       msg.ConnectionId,
		Deposit:            params.QueryDeposit,
		SubmitTimeout:      params.QuerySubmitTimeout,
		RegisteredAtHeight: uint64(ctx.BlockHeader().Height), //nolint:gosec
	}

	m.SetLastRegisteredQueryKey(ctx, lastID)

	if err := m.CollectDeposit(ctx, *registeredQuery); err != nil {
		ctx.Logger().Debug("RegisterInterchainQuery: failed to collect deposit", "message", &msg, "error", err)
		return nil, errors.Wrapf(err, "failed to collect deposit")
	}

	if err := m.SaveQuery(ctx, registeredQuery); err != nil {
		ctx.Logger().Debug("RegisterInterchainQuery: failed to save query", "message", &msg, "error", err)
		return nil, errors.Wrapf(err, "failed to save query: %v", err)
	}

	ctx.EventManager().EmitEvents(getEventsQueryUpdated(registeredQuery))

	return &types.MsgRegisterInterchainQueryResponse{Id: lastID}, nil
}

func (m msgServer) RemoveInterchainQuery(goCtx context.Context, msg *types.MsgRemoveInterchainQueryRequest) (*types.MsgRemoveInterchainQueryResponse, error) {
	if err := msg.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgRemoveInterchainQueryRequest")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.Logger().Debug("RemoveInterchainQuery", "msg", msg)

	query, err := m.GetQueryByID(ctx, msg.GetQueryId())
	if err != nil {
		ctx.Logger().Debug("RemoveInterchainQuery: failed to GetQueryByID",
			"error", err, "query_id", msg.QueryId)
		return nil, errors.Wrapf(err, "failed to get query by query id: %v", err)
	}

	if err := query.ValidateRemoval(ctx, msg.GetSender()); err != nil {
		ctx.Logger().Debug("RemoveInterchainQuery: authorization failed",
			"error", err, "msg", msg)
		return nil, errors.Wrap(sdkerrors.ErrUnauthorized, err.Error())
	}

	m.RemoveQuery(ctx, query)
	m.MustPayOutDeposit(ctx, query.Deposit, msg.GetSigners()[0])
	ctx.EventManager().EmitEvents(getEventsQueryRemoved(query))
	return &types.MsgRemoveInterchainQueryResponse{}, nil
}

func (m msgServer) UpdateInterchainQuery(goCtx context.Context, msg *types.MsgUpdateInterchainQueryRequest) (*types.MsgUpdateInterchainQueryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.Logger().Debug("UpdateInterchainQuery", "msg", msg)
	params := m.GetParams(ctx)

	if err := msg.Validate(params); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgUpdateInterchainQueryRequest")
	}

	query, err := m.GetQueryByID(ctx, msg.GetQueryId())
	if err != nil {
		ctx.Logger().Debug("UpdateInterchainQuery: failed to GetQueryByID",
			"error", err, "query_id", msg.QueryId)
		return nil, errors.Wrapf(err, "failed to get query by query id: %v", err)
	}

	if query.GetOwner() != msg.GetSender() {
		ctx.Logger().Debug("UpdateInterchainQuery: authorization failed",
			"msg", msg)
		return nil, errors.Wrap(sdkerrors.ErrUnauthorized, "authorization failed")
	}

	if err := m.validateUpdateInterchainQueryParams(query, msg); err != nil {
		ctx.Logger().Debug("UpdateInterchainQuery: invalid request",
			"error", err, "query_id", msg.QueryId)
		return nil, errors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if msg.GetNewUpdatePeriod() > 0 {
		query.UpdatePeriod = msg.GetNewUpdatePeriod()
	}
	if len(msg.GetNewKeys()) > 0 && types.InterchainQueryType(query.GetQueryType()).IsKV() {
		query.Keys = msg.GetNewKeys()
	}
	if msg.GetNewTransactionsFilter() != "" && types.InterchainQueryType(query.GetQueryType()).IsTX() {
		query.TransactionsFilter = msg.GetNewTransactionsFilter()
	}

	if err := m.SaveQuery(ctx, query); err != nil {
		ctx.Logger().Debug("UpdateInterchainQuery: failed to save query", "message", &msg, "error", err)
		return nil, errors.Wrapf(err, "failed to save query by query id: %v", err)
	}

	ctx.EventManager().EmitEvents(getEventsQueryUpdated(query))

	return &types.MsgUpdateInterchainQueryResponse{}, nil
}

func (m msgServer) SubmitQueryResult(goCtx context.Context, msg *types.MsgSubmitQueryResult) (*types.MsgSubmitQueryResultResponse, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), LabelRegisterInterchainQuery)

	if err := msg.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgSubmitQueryResult")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.Logger().Debug("SubmitQueryResult", "query_id", msg.QueryId)

	query, err := m.GetQueryByID(ctx, msg.QueryId)
	if err != nil {
		ctx.Logger().Debug("SubmitQueryResult: failed to GetQueryByID",
			"error", err, "query_id", msg.QueryId)
		return nil, errors.Wrapf(err, "failed to get query by id: %v", err)
	}

	connection, ok := m.ibcKeeper.ConnectionKeeper.GetConnection(ctx, query.ConnectionId)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidConnectionID, "registered query %d has invalid connection id: %s", query.Id, query.ConnectionId)
	}

	queryOwner, err := sdk.AccAddressFromBech32(query.Owner)
	if err != nil {
		ctx.Logger().Error("SubmitQueryResult: failed to decode AccAddressFromBech32",
			"error", err, "query", query, "message", msg)
		return nil, errors.Wrapf(err, "failed to decode owner contract address (%s)", query.Owner)
	}

	if msg.Result.KvResults != nil {
		if !types.InterchainQueryType(query.QueryType).IsKV() {
			return nil, errors.Wrapf(types.ErrInvalidType, "invalid query result for query type: %s", query.QueryType)
		}
		if err := m.checkLastRemoteHeight(ctx, *query, ibcclienttypes.NewHeight(msg.Result.Revision, msg.Result.Height)); err != nil {
			return nil, errors.Wrap(types.ErrInvalidHeight, err.Error())
		}
		if len(msg.Result.KvResults) != len(query.Keys) {
			return nil, errors.Wrapf(types.ErrInvalidSubmittedResult, "KV keys length from result is not equal to registered query keys length: %v != %v", len(msg.Result.KvResults), len(query.Keys))
		}

		resp, err := m.ibcKeeper.ConnectionConsensusState(goCtx, &ibcconnectiontypes.QueryConnectionConsensusStateRequest{
			ConnectionId:   query.ConnectionId,
			RevisionNumber: msg.Result.Revision,
			RevisionHeight: msg.Result.Height + 1,
		})
		if err != nil {
			ctx.Logger().Debug("SubmitQueryResult: failed to get ConnectionConsensusState",
				"error", err, "query", query, "message", msg)
			return nil, errors.Wrapf(ibcclienttypes.ErrConsensusStateNotFound, "failed to get consensus state: %v", err)
		}
		consensusStateI, err := ibcclienttypes.UnpackConsensusState(resp.ConsensusState)
		if err != nil {
			ctx.Logger().Error("SubmitQueryResult: failed to UnpackConsensusState",
				"error", err, "query", query, "message", msg)
			return nil, fmt.Errorf("failed marshal: %s, %w", consensusStateI.String(), err)
		}

		consensusState, ok := consensusStateI.(*tendermint.ConsensusState)
		if !ok {
			ctx.Logger().Error("SubmitQueryResult: failed to cast exported.ConsensusState to *tendermint.ConsensusState",
				"error", err, "query", query, "message", msg)
			return nil, errors.Wrapf(sdkerrors.ErrUnpackAny, "failed to cast interface exported.ConsensusState to type *tendermint.ConsensusState")
		}

		clientState, err := m.GetClientState(ctx, connection.ClientId)
		if err != nil {
			return nil, err
		}

		for index, result := range msg.Result.KvResults {
			proof, err := ibccommitmenttypes.ConvertProofs(result.Proof)
			if err != nil {
				ctx.Logger().Debug("SubmitQueryResult: failed to ConvertProofs",
					"error", err, "query", query, "message", msg)
				return nil, errors.Wrapf(types.ErrInvalidType, "failed to convert crypto.ProofOps to MerkleProof: %v", err)
			}

			if !bytes.Equal(result.Key, query.Keys[index].Key) {
				return nil, errors.Wrapf(types.ErrInvalidSubmittedResult, "KV key from result is not equal to registered query key: %v != %v", result.Key, query.Keys[index].Key)
			}

			if result.StoragePrefix != query.Keys[index].Path {
				return nil, errors.Wrapf(types.ErrInvalidSubmittedResult, "KV path from result is not equal to registered query storage prefix: %v != %v", result.StoragePrefix, query.Keys[index].Path)
			}

			path := ibccommitmenttypes.NewMerklePath(result.StoragePrefix, string(result.Key))
			// identify what kind proofs (non-existence proof always has *ics23.CommitmentProof_Nonexist as the first item) we got
			// and call corresponding method to verify it
			switch proof.GetProofs()[0].GetProof().(type) {
			// we can get non-existence proof if someone queried some key which is not exists in the storage on remote chain
			case *ics23.CommitmentProof_Nonexist:
				if err := proof.VerifyNonMembership(clientState.ProofSpecs, consensusState.GetRoot(), path); err != nil {
					ctx.Logger().Debug("SubmitQueryResult: failed to VerifyNonMembership",
						"error", err, "query", query, "message", msg, "path", path)
					return nil, errors.Wrapf(types.ErrInvalidProof, "failed to verify proof: %v", err)
				}
				result.Value = nil
			case *ics23.CommitmentProof_Exist:
				if err := proof.VerifyMembership(clientState.ProofSpecs, consensusState.GetRoot(), path, result.Value); err != nil {
					ctx.Logger().Debug("SubmitQueryResult: failed to VerifyMembership",
						"error", err, "query", query, "message", msg, "path", path)
					return nil, errors.Wrapf(types.ErrInvalidProof, "failed to verify proof: %v", err)
				}
			default:
				return nil, errors.Wrapf(types.ErrInvalidProof, "unknown proof type %T", proof.GetProofs()[0].GetProof())
			}
		}

		if err = m.saveKVQueryResult(ctx, query, msg.Result); err != nil {
			ctx.Logger().Error("SubmitQueryResult: failed to SaveKVQueryResult",
				"error", err, "query", query, "message", msg)
			return nil, errors.Wrapf(err, "failed to SaveKVQueryResult: %v", err)
		}

		if msg.Result.GetAllowKvCallbacks() {
			// Let the query owner contract process the query result.
			if _, err := m.contractManagerKeeper.SudoKVQueryResult(ctx, queryOwner, query.Id); err != nil {
				ctx.Logger().Debug("SubmitQueryResult: failed to SudoKVQueryResult",
					"error", err, "query_id", query.GetId())
				return nil, errors.Wrapf(err, "contract %s rejected KV query result (query_id: %d)",
					queryOwner, query.GetId())
			}
			return &types.MsgSubmitQueryResultResponse{}, nil
		}
	}

	if msg.Result.Block != nil && msg.Result.Block.Tx != nil {
		if !types.InterchainQueryType(query.QueryType).IsTX() {
			return nil, errors.Wrapf(types.ErrInvalidType, "invalid query result for query type: %s", query.QueryType)
		}

		if err := m.ProcessBlock(ctx, queryOwner, msg.QueryId, connection.ClientId, msg.Result.Block); err != nil {
			ctx.Logger().Debug("SubmitQueryResult: failed to ProcessBlock",
				"error", err, "query", query, "message", msg)
			return nil, errors.Wrapf(err, "failed to ProcessBlock: %v", err)
		}

		if err = m.UpdateLastLocalHeight(ctx, query.Id, uint64(ctx.BlockHeight())); err != nil { //nolint:gosec
			return nil, errors.Wrapf(err,
				"failed to update last local height for a result with id %d: %v", query.Id, err)
		}
	}

	return &types.MsgSubmitQueryResultResponse{}, nil
}

// validateUpdateInterchainQueryParams checks whether the parameters to be updated corresponds
// with the query type.
func (m msgServer) validateUpdateInterchainQueryParams(
	query *types.RegisteredQuery,
	msg *types.MsgUpdateInterchainQueryRequest,
) error {
	queryType := types.InterchainQueryType(query.GetQueryType())
	newKvKeysSet := len(msg.GetNewKeys()) != 0
	newTxFilterSet := msg.GetNewTransactionsFilter() != ""

	if queryType.IsKV() && !newKvKeysSet && newTxFilterSet {
		return fmt.Errorf("params to update don't correspond with query type: can't update TX filter for a KV query")
	}
	if queryType.IsTX() && !newTxFilterSet && newKvKeysSet {
		return fmt.Errorf("params to update don't correspond with query type: can't update KV keys for a TX query")
	}
	return nil
}

// UpdateParams updates the module parameters
func (k Keeper) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgUpdateParams")
	}

	authority := k.GetAuthority()
	if authority != req.Authority {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid authority; expected %s, got %s", authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func getEventsQueryUpdated(query *types.RegisteredQuery) sdk.Events {
	return sdk.Events{
		sdk.NewEvent(
			types.EventTypeNeutronMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueQueryUpdated),
			sdk.NewAttribute(types.AttributeKeyQueryID, strconv.FormatUint(query.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyConnectionID, query.ConnectionId),
			sdk.NewAttribute(types.AttributeKeyOwner, query.Owner),
			sdk.NewAttribute(types.AttributeKeyQueryType, query.QueryType),
			sdk.NewAttribute(types.AttributeTransactionsFilterQuery, query.TransactionsFilter),
			sdk.NewAttribute(types.AttributeKeyKVQuery, types.KVKeys(query.Keys).String()),
		),
	}
}

func getEventsQueryRemoved(query *types.RegisteredQuery) sdk.Events {
	return sdk.Events{
		sdk.NewEvent(
			types.EventTypeNeutronMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueQueryRemoved),
			sdk.NewAttribute(types.AttributeKeyQueryID, strconv.FormatUint(query.Id, 10)),
			sdk.NewAttribute(types.AttributeKeyConnectionID, query.ConnectionId),
			sdk.NewAttribute(types.AttributeKeyOwner, query.Owner),
			sdk.NewAttribute(types.AttributeKeyQueryType, query.QueryType),
			sdk.NewAttribute(types.AttributeTransactionsFilterQuery, query.TransactionsFilter),
			sdk.NewAttribute(types.AttributeKeyKVQuery, types.KVKeys(query.Keys).String()),
		),
	}
}
