package keeper

import (
	"context"
	"fmt"
	"strconv"

	"github.com/neutron-org/neutron/x/incentives/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.MsgServer = msgServer{}

// msgServer provides a way to reference keeper pointer in the message server interface.
type msgServer struct {
	keeper *Keeper
}

// NewMsgServerImpl returns an instance of MsgServer for the provided keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

// CreateGauge creates a gauge and sends coins to the gauge.
// Emits create gauge event and returns the create gauge response.
func (server msgServer) CreateGauge(
	goCtx context.Context,
	msg *types.MsgCreateGauge,
) (*types.MsgCreateGaugeResponse, error) {
	if server.keeper.authority != msg.Owner {
		return nil, sdkerrors.Wrapf(
			types.ErrInvalidSigner,
			"invalid authority; expected %s, got %s",
			server.keeper.authority,
			msg.Owner,
		)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	gauge, err := server.keeper.CreateGauge(
		ctx,
		msg.IsPerpetual,
		owner,
		msg.Coins,
		msg.DistributeTo,
		msg.StartTime,
		msg.NumEpochsPaidOver,
		msg.PricingTick,
	)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtCreateGauge,
			sdk.NewAttribute(types.AttributeGaugeID, strconv.FormatUint(gauge.Id, 10)),
		),
	})

	return &types.MsgCreateGaugeResponse{}, nil
}

// AddToGauge adds coins to gauge.
// Emits add to gauge event and returns the add to gauge response.
func (server msgServer) AddToGauge(
	goCtx context.Context,
	msg *types.MsgAddToGauge,
) (*types.MsgAddToGaugeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	err = server.keeper.AddToGaugeRewards(ctx, owner, msg.Rewards, msg.GaugeId)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtAddToGauge,
			sdk.NewAttribute(types.AttributeGaugeID, strconv.FormatUint(msg.GaugeId, 10)),
		),
	})

	return &types.MsgAddToGaugeResponse{}, nil
}

// StakeTokens stakes tokens in either two ways.
// 1. Add to an existing stake if a stake with the same owner and same duration exists.
// 2. Create a new stake if not.
// A sanity check to ensure given tokens is a single token is done in ValidateBaic.
// That is, a stake with multiple tokens cannot be created.
func (server msgServer) Stake(
	goCtx context.Context,
	msg *types.MsgStake,
) (*types.MsgStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	params := server.keeper.GetParams(ctx)
	startDistEpoch := server.keeper.ek.GetEpochInfo(ctx, params.DistrEpochIdentifier).CurrentEpoch

	// if the owner + duration combination is new, create a new stake.
	stake, err := server.keeper.CreateStake(ctx, owner, msg.Coins, ctx.BlockTime(), startDistEpoch)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtStake,
			sdk.NewAttribute(types.AttributeStakeID, strconv.FormatUint(stake.ID, 10)),
			sdk.NewAttribute(types.AttributeStakeOwner, stake.Owner),
			sdk.NewAttribute(types.AttributeStakeAmount, stake.Coins.String()),
		),
	})

	return &types.MsgStakeResponse{ID: stake.ID}, nil
}

// BeginUnstaking begins unstaking of the specified stake.
// The stake would enter the unstaking queue, with the endtime of the stake set as block time + duration.
func (server msgServer) Unstake(
	goCtx context.Context,
	msg *types.MsgUnstake,
) (*types.MsgUnstakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	unstakes := msg.Unstakes
	if len(msg.Unstakes) == 0 {
		stakes := server.keeper.GetStakesByAccount(ctx, sdk.AccAddress(msg.Owner))
		unstakes = make([]*types.MsgUnstake_UnstakeDescriptor, len(stakes))
		for i, stake := range stakes {
			unstakes[i] = &types.MsgUnstake_UnstakeDescriptor{
				ID:    stake.ID,
				Coins: sdk.NewCoins(),
			}
		}
	}

	for _, unstake := range unstakes {
		stake, err := server.keeper.GetStakeByID(ctx, unstake.ID)
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrInvalidRequest, err.Error())
		}

		if msg.Owner != stake.Owner {
			return nil, sdkerrors.Wrap(
				types.ErrNotStakeOwner,
				fmt.Sprintf(
					"msg sender (%s) and stake owner (%s) does not match",
					msg.Owner,
					stake.Owner,
				),
			)
		}

		_, err = server.keeper.Unstake(ctx, stake, unstake.Coins)
		if err != nil {
			return nil, sdkerrors.Wrap(types.ErrInvalidRequest, err.Error())
		}
	}

	// N.B. begin unstake event is emitted downstream in the keeper method.
	return &types.MsgUnstakeResponse{}, nil
}

func (server msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if err := req.ValidateBasic(); err != nil {
		return nil, err
	}
	authority := server.keeper.GetAuthority()
	if authority != req.Authority {
		return nil, sdkerrors.Wrapf(types.ErrInvalidRequest, "invalid authority; expected %s, got %s", authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := server.keeper.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
