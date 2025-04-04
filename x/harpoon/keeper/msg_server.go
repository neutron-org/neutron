package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/harpoon/types"
)

type msgServer struct {
	keeper *Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// ManageHookSubscription updates the hook subscriptions for a specified contract address.
// Can only be executed by the module's authority.
func (s msgServer) ManageHookSubscription(goCtx context.Context, req *types.MsgManageHookSubscription) (*types.MsgManageHookSubscriptionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, errorsmod.Wrapf(err, "failed to validate manage hook subscription message")
	}

	contractAddr := sdk.MustAccAddressFromBech32(req.HookSubscription.ContractAddress)
	if !s.keeper.wasmKeeper.HasContractInfo(goCtx, contractAddr) {
		return nil, errorsmod.Wrapf(errors.ErrInvalidAddress, "contract address not found: %s", contractAddr)
	}

	if s.keeper.GetAuthority() != req.Authority {
		return nil, errorsmod.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", s.keeper.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	s.keeper.UpdateHookSubscription(ctx, req.HookSubscription)

	return &types.MsgManageHookSubscriptionResponse{}, nil
}
