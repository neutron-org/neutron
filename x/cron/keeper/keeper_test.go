package keeper_test

import (
	"fmt"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/neutron-org/neutron/testutil"
	testutil_keeper "github.com/neutron-org/neutron/testutil/cron/keeper"
	mock_types "github.com/neutron-org/neutron/testutil/mocks/cron/types"
	"github.com/neutron-org/neutron/x/cron/types"
	"github.com/stretchr/testify/require"
)

// ExecuteReadySchedules
// - calls msgServer.execute() on ready schedules
// - updates ready schedules executeHeight
// - does not update heights of unready schedules
// - does not go over the limit
func TestKeeperExecuteReadySchedules(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accountKeeper := mock_types.NewMockAccountKeeper(ctrl)
	addr, err := sdk.AccAddressFromBech32(testutil.TestOwnerAddress)
	require.NoError(t, err)

	wasmMsgServer := mock_types.NewMockWasmMsgServer(ctrl)
	k, ctx := testutil_keeper.CronKeeper(t, wasmMsgServer, accountKeeper)
	ctx = ctx.WithBlockHeight(0)

	k.SetParams(ctx, types.Params{
		AdminAddress:    testutil.TestOwnerAddress,
		SecurityAddress: testutil.TestOwnerAddress,
		Limit:           2,
	})

	schedules := []types.Schedule{
		{
			Name:   "1_unready1",
			Period: 3,
			Msgs: []types.MsgExecuteContract{
				{
					Contract: "1_neutron",
					Msg:      []byte("1_msg"),
				},
			},
			LastExecuteHeight: 4,
		},
		{
			Name:   "2_ready1",
			Period: 3,
			Msgs: []types.MsgExecuteContract{
				{
					Contract: "2_neutron",
					Msg:      []byte("2_msg"),
				},
			},
			LastExecuteHeight: 0,
		},
		{
			Name:   "3_ready2",
			Period: 3,
			Msgs: []types.MsgExecuteContract{
				{
					Contract: "3_neutron",
					Msg:      []byte("3_msg"),
				},
			},
			LastExecuteHeight: 0,
		},
		{
			Name:              "4_unready2",
			Period:            3,
			Msgs:              []types.MsgExecuteContract{},
			LastExecuteHeight: 4,
		},
		{
			Name:   "5_ready3",
			Period: 3,
			Msgs: []types.MsgExecuteContract{
				{
					Contract: "5_neutron",
					Msg:      []byte("5_msg"),
				},
			},
			LastExecuteHeight: 0,
		},
	}

	for _, item := range schedules {
		ctx = ctx.WithBlockHeight(int64(item.LastExecuteHeight))
		err := k.AddSchedule(ctx, item.Name, item.Period, item.Msgs)
		require.NoError(t, err)
	}

	ctx = ctx.WithBlockHeight(5)

	accountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(addr)
	accountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(addr)
	wasmMsgServer.EXPECT().ExecuteContract(sdk.WrapSDKContext(ctx), &wasmtypes.MsgExecuteContract{
		Sender:   testutil.TestOwnerAddress,
		Contract: "2_neutron",
		Msg:      []byte("2_msg"),
		Funds:    sdk.NewCoins(),
	}).Return(nil, fmt.Errorf("executeerror"))
	wasmMsgServer.EXPECT().ExecuteContract(sdk.WrapSDKContext(ctx), &wasmtypes.MsgExecuteContract{
		Sender:   testutil.TestOwnerAddress,
		Contract: "3_neutron",
		Msg:      []byte("3_msg"),
		Funds:    sdk.NewCoins(),
	}).Return(&wasmtypes.MsgExecuteContractResponse{}, nil)

	k.ExecuteReadySchedules(ctx)

	unready1, _ := k.GetSchedule(ctx, "1_unready1")
	ready1, _ := k.GetSchedule(ctx, "2_ready1")
	ready2, _ := k.GetSchedule(ctx, "3_ready2")
	unready2, _ := k.GetSchedule(ctx, "4_unready2")
	ready3, _ := k.GetSchedule(ctx, "5_ready3")

	require.Equal(t, uint64(4), unready1.LastExecuteHeight)
	require.Equal(t, uint64(5), ready1.LastExecuteHeight)
	require.Equal(t, uint64(5), ready2.LastExecuteHeight)
	require.Equal(t, uint64(4), unready2.LastExecuteHeight)
	require.Equal(t, uint64(0), ready3.LastExecuteHeight)

	// let's make another call at the next height
	// Notice that now only one ready schedule left because we got limit of 2 at once
	ctx = ctx.WithBlockHeight(6)

	accountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(addr)
	wasmMsgServer.EXPECT().ExecuteContract(sdk.WrapSDKContext(ctx), &wasmtypes.MsgExecuteContract{
		Sender:   testutil.TestOwnerAddress,
		Contract: "5_neutron",
		Msg:      []byte("5_msg"),
		Funds:    sdk.NewCoins(),
	}).Return(&wasmtypes.MsgExecuteContractResponse{}, nil)

	k.ExecuteReadySchedules(ctx)

	unready1, _ = k.GetSchedule(ctx, "1_unready1")
	ready1, _ = k.GetSchedule(ctx, "2_ready1")
	ready2, _ = k.GetSchedule(ctx, "3_ready2")
	unready2, _ = k.GetSchedule(ctx, "4_unready2")
	ready3, _ = k.GetSchedule(ctx, "5_ready3")

	require.Equal(t, uint64(4), unready1.LastExecuteHeight)
	require.Equal(t, uint64(5), ready1.LastExecuteHeight)
	require.Equal(t, uint64(5), ready2.LastExecuteHeight)
	require.Equal(t, uint64(4), unready2.LastExecuteHeight)
	require.Equal(t, uint64(6), ready3.LastExecuteHeight)
}

// AddSchedule
// - adds new schedule if ok
// - returns error if exists

// RemoveSchedule
// - removes schedule
// - if not found does not fail

// GetSchedule gets schedule or not found

// GetAllSchedules gets all schedules
