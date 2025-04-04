package keeper_test

import (
	"context"
	"testing"

	types2 "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/neutron-org/neutron/v6/testutil"

	"github.com/stretchr/testify/require"

	keepertest "github.com/neutron-org/neutron/v6/testutil/harpoon/keeper"
	"github.com/neutron-org/neutron/v6/x/harpoon/keeper"
	"github.com/neutron-org/neutron/v6/x/harpoon/types"

	"github.com/golang/mock/gomock"

	mock_types "github.com/neutron-org/neutron/v6/testutil/mocks/harpoon/types"
)

const (
	ContractAddress1 = "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a"
	ContractAddress2 = "neutron1u9dulasrfe6laxwwjx83njhm5as466uz43arfheq79k8eqahqhasf3tj94"
	ContractAddress3 = "neutron1k5996rnrjcu4dwxtjq46zuh4dryl6pdrlyja7fw2qm7yarkvk97s9uzmuz"
)

func setupMsgServer(t testing.TB) (keeper.Keeper, types.MsgServer, context.Context) {
	k, ctx := keepertest.HarpoonKeeper(t, nil)
	return *k, keeper.NewMsgServerImpl(k), ctx
}

func TestMsgServer(t *testing.T) {
	k, ms, ctx := setupMsgServer(t)
	require.NotNil(t, ms)
	require.NotNil(t, ctx)
	require.NotEmpty(t, k)
}

func TestManageHookSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name                      string
		manageHookSubscriptionMsg types.MsgManageHookSubscription
		malleate                  func(ctx context.Context, mockWasmKeeper *mock_types.MockWasmKeeper)
		expectedErr               string
	}{
		{
			"empty authority",
			types.MsgManageHookSubscription{
				Authority: "",
				HookSubscription: &types.HookSubscription{
					ContractAddress: testutil.TestOwnerAddress,
					Hooks:           []types.HookType{},
				},
			},
			func(_ context.Context, _ *mock_types.MockWasmKeeper) {
			},
			"authority is invalid: empty address string is not allowed",
		},
		{
			"non unique hooks",
			types.MsgManageHookSubscription{
				Authority: "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z",
				HookSubscription: &types.HookSubscription{
					ContractAddress: testutil.TestOwnerAddress,
					Hooks:           []types.HookType{types.HOOK_TYPE_AFTER_VALIDATOR_BONDED, types.HOOK_TYPE_AFTER_DELEGATION_MODIFIED, types.HOOK_TYPE_AFTER_VALIDATOR_BONDED},
				},
			},
			func(_ context.Context, _ *mock_types.MockWasmKeeper) {
			},
			"subscription hooks are not unique",
		},
		{
			"non existing hook type",
			types.MsgManageHookSubscription{
				Authority: "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z",
				HookSubscription: &types.HookSubscription{
					ContractAddress: testutil.TestOwnerAddress,
					Hooks:           []types.HookType{types.HookType(100)},
				},
			},
			func(_ context.Context, _ *mock_types.MockWasmKeeper) {
			},
			"non-existing hook type",
		},
		{
			"bad case - non-existing contract",
			types.MsgManageHookSubscription{
				Authority: "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z",
				HookSubscription: &types.HookSubscription{
					ContractAddress: "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z",
					Hooks:           []types.HookType{types.HOOK_TYPE_AFTER_VALIDATOR_CREATED, types.HOOK_TYPE_AFTER_DELEGATION_MODIFIED},
				},
			},
			func(ctx context.Context, mockWasmKeeper *mock_types.MockWasmKeeper) {
				mockWasmKeeper.EXPECT().HasContractInfo(ctx, types2.MustAccAddressFromBech32("neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z")).Return(false)
			},
			errors.ErrInvalidAddress.Error(),
		},
		{
			"good case - empty hooks",
			types.MsgManageHookSubscription{
				Authority: "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z",
				HookSubscription: &types.HookSubscription{
					ContractAddress: testutil.TestOwnerAddress,
					Hooks:           []types.HookType{},
				},
			},
			func(ctx context.Context, mockWasmKeeper *mock_types.MockWasmKeeper) {
				mockWasmKeeper.EXPECT().HasContractInfo(ctx, types2.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
			},
			"",
		},
		{
			"good case - some hooks present",
			types.MsgManageHookSubscription{
				Authority: "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z",
				HookSubscription: &types.HookSubscription{
					ContractAddress: testutil.TestOwnerAddress,
					Hooks:           []types.HookType{types.HOOK_TYPE_AFTER_VALIDATOR_CREATED, types.HOOK_TYPE_AFTER_DELEGATION_MODIFIED},
				},
			},
			func(ctx context.Context, mockWasmKeeper *mock_types.MockWasmKeeper) {
				mockWasmKeeper.EXPECT().HasContractInfo(ctx, types2.MustAccAddressFromBech32(testutil.TestOwnerAddress)).Return(true)
			},
			"",
		},
	}

	for _, tt := range tests {
		wasmKeeper := mock_types.NewMockWasmKeeper(ctrl)

		k, ctx := keepertest.HarpoonKeeper(t, wasmKeeper)
		msgServer := keeper.NewMsgServerImpl(k)
		tt.malleate(ctx, wasmKeeper)

		res, err := msgServer.ManageHookSubscription(ctx, &tt.manageHookSubscriptionMsg)

		if tt.expectedErr == "" {
			require.NoError(t, err, tt.expectedErr)
			require.Equal(t, res, &types.MsgManageHookSubscriptionResponse{})
		} else {
			require.ErrorContains(t, err, tt.expectedErr)
			require.Empty(t, res)
		}
	}
}
