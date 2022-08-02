package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	"github.com/neutron-org/neutron/x/interchaintxs/types"
)

func (k Keeper) InterchainAccountAddress(c context.Context, req *types.QueryInterchainAccountAddressRequest) (*types.QueryInterchainAccountAddressResponse, error) {
	if req == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	
	icaOwner, err := types.NewICAOwner(req.OwnerAddress, req.InterchainAccountId)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "could not find account: %s", err)
	}

	portID, err := icatypes.NewControllerPortID(icaOwner.String())
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "could not find account: %s", err)
	}

	addr, found := k.icaControllerKeeper.GetInterchainAccountAddress(ctx, req.ConnectionId, portID)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrInterchainAccountNotFound, "no account found for portID %s", portID)
	}

	return &types.QueryInterchainAccountAddressResponse{InterchainAccountAddress: addr}, nil
}
