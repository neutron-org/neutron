package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v5/x/harpoon/types"
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

func (k msgServer) ManageHookSubscription(goCtx context.Context, req *types.MsgManageHookSubscription) (*types.MsgManageHookSubscriptionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errorsmod.Wrapf(err, "failed to validate manage hook subscription message")
	}

	if k.GetAuthority() != req.Authority {
		return nil, errorsmod.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.UpdateHookSubscription(ctx, req.HookSubscription)

	return &types.MsgManageHookSubscriptionResponse{}, nil
}
