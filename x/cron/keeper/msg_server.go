package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/neutron-org/neutron/v4/x/cron/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// AddSchedule adds new schedule
func (k msgServer) AddSchedule(goCtx context.Context, req *types.MsgAddSchedule) (*types.MsgAddScheduleResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgAddSchedule")
	}

	authority, reqAuthority := k.GetAuthority(), req.GetAuthority()
	if authority != reqAuthority {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid authority; expected %s, got %s", authority, reqAuthority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.Keeper.AddSchedule(ctx, req.GetName(), req.GetPeriod(), req.GetMsgs(), uint64(req.GetBlocker())); err != nil {
		return nil, err
	}

	return &types.MsgAddScheduleResponse{}, nil
}

// RemoveSchedule removes schedule
func (k msgServer) RemoveSchedule(goCtx context.Context, req *types.MsgRemoveSchedule) (*types.MsgRemoveScheduleResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgRemoveSchedule")
	}

	authority, reqAuthority := k.GetAuthority(), req.GetAuthority()
	if authority != reqAuthority {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid authority; expected %s, got %s", authority, reqAuthority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Keeper.RemoveSchedule(ctx, req.GetName())

	return &types.MsgRemoveScheduleResponse{}, nil
}

// UpdateParams updates the module parameters
func (k msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate MsgUpdateParams")
	}

	authority, reqAuthority := k.GetAuthority(), req.GetAuthority()
	if authority != reqAuthority {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid authority; expected %s, got %s", authority, reqAuthority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
