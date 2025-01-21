package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/v5/x/harpoon/types"
	"google.golang.org/grpc/codes"

	"context"

	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) SubscribedContracts(goCtx context.Context, req *types.QuerySubscribedContracts) (*types.QuerySubscribedContractsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.HookType == "" {
		return nil, status.Error(codes.InvalidArgument, "empty hook type")
	}

	if hookTypeInt, ok := types.HookType_value[req.HookType]; ok {
		ctx := sdk.UnwrapSDKContext(goCtx)
		hookType := types.HookType(hookTypeInt)
		return &types.QuerySubscribedContractsResponse{ContractAddresses: k.GetSubscribedAddressesForHookType(ctx, hookType)}, nil
	} else {
		return nil, status.Error(codes.InvalidArgument, "non existing hookType")
	}
}
