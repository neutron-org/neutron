package keeper

import (
	"context"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"

	"github.com/neutron-org/neutron/v6/x/interchaintxs/types"
)

func (k Keeper) InterchainAccountAddress(c context.Context, req *types.QueryInterchainAccountAddressRequest) (*types.QueryInterchainAccountAddressResponse, error) {
	if req == nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	icaOwner, err := types.NewICAOwner(req.OwnerAddress, req.InterchainAccountId)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "failed to create ica owner: %s", err)
	}

	portID, err := icatypes.NewControllerPortID(icaOwner.String())
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "failed to get controller portID: %s", err)
	}

	addr, found := k.icaControllerKeeper.GetInterchainAccountAddress(ctx, req.ConnectionId, portID)
	if !found {
		return nil, errors.Wrapf(types.ErrInterchainAccountNotFound, "no interchain account found for portID %s", portID)
	}

	return &types.QueryInterchainAccountAddressResponse{InterchainAccountAddress: addr}, nil
}
