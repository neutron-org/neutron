package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v5/x/harpoon/types"
)

func (k msgServer) ManageHookSubscription(goCtx context.Context, req *types.MsgManageHookSubscription) (*types.MsgManageHookSubscriptionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errorsmod.Wrapf(err, "failed to validate manage hook subscription message")
	}

	if k.GetAuthority() != req.Authority {
		return nil, errorsmod.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.UpdateHookSubscription(ctx, req.HookSubscription); err != nil {
		return nil, err
	}

	return &types.MsgManageHookSubscriptionResponse{}, nil
}
