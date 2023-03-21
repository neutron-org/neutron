package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	types2 "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/testutil"
	testkeeper "github.com/neutron-org/neutron/testutil/interchaintxs/keeper"
	mock_types "github.com/neutron-org/neutron/testutil/mocks/interchaintxs/types"
	"github.com/neutron-org/neutron/x/interchaintxs/types"
)

func TestKeeper_InterchainAccountAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	icaKeeper := mock_types.NewMockICAControllerKeeper(ctrl)
	keeper, ctx := testkeeper.InterchainTxsKeeper(t, nil, nil, icaKeeper, nil, nil)
	wctx := sdk.WrapSDKContext(ctx)

	resp, err := keeper.InterchainAccountAddress(wctx, nil)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
	require.Nil(t, resp)

	resp, err = keeper.InterchainAccountAddress(wctx, &types.QueryInterchainAccountAddressRequest{
		OwnerAddress:        "nonbetch32",
		InterchainAccountId: "test1",
		ConnectionId:        "connection-0",
	})
	require.ErrorContains(t, err, "failed to create ica owner")
	require.Nil(t, resp)

	portID := fmt.Sprintf("%s%s.%s", types2.PortPrefix, testutil.TestOwnerAddress, "test1")
	icaKeeper.EXPECT().GetInterchainAccountAddress(ctx, "connection-0", portID).Return("", false)
	resp, err = keeper.InterchainAccountAddress(wctx, &types.QueryInterchainAccountAddressRequest{
		OwnerAddress:        testutil.TestOwnerAddress,
		InterchainAccountId: "test1",
		ConnectionId:        "connection-0",
	})
	require.ErrorContains(t, err, "no interchain account found for portID")
	require.Nil(t, resp)

	portID = fmt.Sprintf("%s%s.%s", types2.PortPrefix, testutil.TestOwnerAddress, "test1")
	icaKeeper.EXPECT().GetInterchainAccountAddress(ctx, "connection-0", portID).Return("neutron1interchainaccountaddress", true)
	resp, err = keeper.InterchainAccountAddress(wctx, &types.QueryInterchainAccountAddressRequest{
		OwnerAddress:        testutil.TestOwnerAddress,
		InterchainAccountId: "test1",
		ConnectionId:        "connection-0",
	})
	require.NoError(t, err)
	require.Equal(t, &types.QueryInterchainAccountAddressResponse{InterchainAccountAddress: "neutron1interchainaccountaddress"}, resp)
}
