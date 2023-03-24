package cron_test

//TODO: FIXME
//func TestGenesis(t *testing.T) {
//	genesisState := types.GenesisState{
//		Params: types.DefaultParams(),
//
//		ScheduleList: []types.Schedule{
//		{
//			Index: "0",
//},
//		{
//			Index: "1",
//},
//	},
//	// this line is used by starport scaffolding # genesis/test/state
//	}
//
//	k, ctx := keepertest.CronKeeper(t)
//	cron.InitGenesis(ctx, *k, genesisState)
//	got := cron.ExportGenesis(ctx, *k)
//	require.NotNil(t, got)
//
//	nullify.Fill(&genesisState)
//	nullify.Fill(got)
//
//	require.ElementsMatch(t, genesisState.ScheduleList, got.ScheduleList)
//// this line is used by starport scaffolding # genesis/test/assert
//}
