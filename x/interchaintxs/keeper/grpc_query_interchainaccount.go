package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	"github.com/neutron-org/neutron/x/interchaintxs/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) InterchainAccountAddress(c context.Context, req *types.QueryInterchainAccountAddressRequest) (*types.QueryInterchainAccountAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	portID, err := icatypes.NewControllerPortID(req.OwnerAddress)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not find account: %s", err)
	}

	addr, found := k.icaControllerKeeper.GetInterchainAccountAddress(ctx, req.ConnectionId, portID)
	if !found {
		return nil, status.Errorf(codes.NotFound, "no account found for portID %s", portID)
	}

	return &types.QueryInterchainAccountAddressResponse{InterchainAccountAddress: addr}, nil
}
