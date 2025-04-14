package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"

	"github.com/neutron-org/neutron/v6/x/harpoon/types"

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

	if err := types.ValidateHookType(req.HookType); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.QuerySubscribedContractsResponse{ContractAddresses: s.keeper.GetSubscribedAddressesForHookType(ctx, req.HookType)}, nil
}
