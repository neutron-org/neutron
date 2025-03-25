package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	revenuetypes "github.com/neutron-org/neutron/v6/x/revenue/types"
)

type msgServer struct {
	keeper *Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) revenuetypes.MsgServer {
	return &msgServer{keeper: keeper}
}

var _ revenuetypes.MsgServer = msgServer{}

func (s msgServer) UpdateParams(goCtx context.Context, msg *revenuetypes.MsgUpdateParams) (*revenuetypes.MsgUpdateParamsResponse, error) {
	if err := msg.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid MsgUpdateParams")
	}

	authority := s.keeper.GetAuthority()
	if authority != msg.Authority {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid authority; expected %s, got %s", authority, msg.Authority)
	}

	ctx := sdktypes.UnwrapSDKContext(goCtx)
	if err := s.keeper.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &revenuetypes.MsgUpdateParamsResponse{}, nil
}

func (s msgServer) FundTreasury(goCtx context.Context, msg *revenuetypes.MsgFundTreasury) (*revenuetypes.MsgFundTreasuryResponse, error) {
	if err := msg.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid MsgFundTreasury")
	}

	sender, err := sdktypes.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create acc address from bech32 %s: %s", msg.Sender, err)
	}

	ctx := sdktypes.UnwrapSDKContext(goCtx)
	if err := s.keeper.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		sender,
		revenuetypes.RevenueTreasuryPoolName,
		sdktypes.NewCoins(sdktypes.NewCoin(
			msg.Amount[0].Denom, msg.Amount[0].Amount,
		)),
	); err != nil {
		return nil, err
	}

	return &revenuetypes.MsgFundTreasuryResponse{}, nil
}
