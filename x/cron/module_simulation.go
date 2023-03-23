package cron

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	cronsimulation "github.com/neutron-org/neutron/x/cron/simulation"
	"github.com/neutron-org/neutron/x/cron/types"
)

// avoid unused import issue
var (
	_ = cronsimulation.FindAccount
	_ = simappparams.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
opWeightMsgCreateSchedule = "op_weight_msg_schedule"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCreateSchedule int = 100

	opWeightMsgUpdateSchedule = "op_weight_msg_schedule"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdateSchedule int = 100

	opWeightMsgDeleteSchedule = "op_weight_msg_schedule"
	// TODO: Determine the simulation weight value
	defaultWeightMsgDeleteSchedule int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	cronGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		ScheduleList: []types.Schedule{
		{
			Creator: sample.AccAddress(),
Index: "0",
},
		{
			Creator: sample.AccAddress(),
Index: "1",
},
	},
	// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&cronGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized  param changes for the simulator
func (am AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{}
}

// RegisterStoreDecoder registers a decoder
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgCreateSchedule int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgCreateSchedule, &weightMsgCreateSchedule, nil,
		func(_ *rand.Rand) {
			weightMsgCreateSchedule = defaultWeightMsgCreateSchedule
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateSchedule,
		cronsimulation.SimulateMsgCreateSchedule(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateSchedule int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgUpdateSchedule, &weightMsgUpdateSchedule, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateSchedule = defaultWeightMsgUpdateSchedule
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateSchedule,
		cronsimulation.SimulateMsgUpdateSchedule(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgDeleteSchedule int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgDeleteSchedule, &weightMsgDeleteSchedule, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteSchedule = defaultWeightMsgDeleteSchedule
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeleteSchedule,
		cronsimulation.SimulateMsgDeleteSchedule(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
