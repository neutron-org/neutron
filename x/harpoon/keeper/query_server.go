package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"

	"github.com/neutron-org/neutron/v5/x/harpoon/types"

	"google.golang.org/grpc/status"
)

type queryServer struct {
	keeper *Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper *Keeper) types.QueryServer {
	return &queryServer{keeper: keeper}
}

var _ types.QueryServer = queryServer{}

// SubscribedContracts retrieves the contract addresses subscribed to a specific hook type.
func (s queryServer) SubscribedContracts(goCtx context.Context, req *types.QuerySubscribedContractsRequest) (*types.QuerySubscribedContractsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.HookType == "" {
		return nil, status.Error(codes.InvalidArgument, "empty hook type")
	}

	if hookTypeInt, ok := types.HookType_value[req.HookType]; ok {
		ctx := sdk.UnwrapSDKContext(goCtx)
		hookType := types.HookType(hookTypeInt)
		return &types.QuerySubscribedContractsResponse{ContractAddresses: s.keeper.GetSubscribedAddressesForHookType(ctx, hookType)}, nil
	} else {
		return nil, status.Error(codes.InvalidArgument, "non existing hookType")
	}
}
