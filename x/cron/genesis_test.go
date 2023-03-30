package cron_test

import (
	"testing"

	"github.com/neutron-org/neutron/testutil/cron/keeper"
	"github.com/neutron-org/neutron/testutil/cron/nullify"
	"github.com/neutron-org/neutron/x/cron"
	"github.com/neutron-org/neutron/x/cron/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		ScheduleList: []types.Schedule{
			{
				Name:              "a",
				Period:            5,
				Msgs:              nil,
				LastExecuteHeight: 0,
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keeper.CronKeeper(t, nil, nil)
	cron.InitGenesis(ctx, *k, genesisState)
	got := cron.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.Params, got.Params)
	require.ElementsMatch(t, genesisState.ScheduleList, got.ScheduleList)
	// this line is used by starport scaffolding # genesis/test/assert
}
