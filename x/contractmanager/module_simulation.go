package contractmanager

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/neutron-org/neutron/testutil/contractmanager/sample"
	contractmanagersimulation "github.com/neutron-org/neutron/x/contractmanager/simulation"
	"github.com/neutron-org/neutron/x/contractmanager/types"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = contractmanagersimulation.FindAccount
	_ = simappparams.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	opWeightMsgCreateFailure = "op_weight_msg_failure"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCreateFailure int = 100

	opWeightMsgUpdateFailure = "op_weight_msg_failure"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdateFailure int = 100

	opWeightMsgDeleteFailure = "op_weight_msg_failure"
	// TODO: Determine the simulation weight value
	defaultWeightMsgDeleteFailure int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	contractmanagerGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		FailureList: []types.Failure{
			{
				Creator: sample.AccAddress(),
				Index:   "0",
			},
			{
				Creator: sample.AccAddress(),
				Index:   "1",
			},
		},
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&contractmanagerGenesis)
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
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgCreateFailure int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgCreateFailure, &weightMsgCreateFailure, nil,
		func(_ *rand.Rand) {
			weightMsgCreateFailure = defaultWeightMsgCreateFailure
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateFailure,
		contractmanagersimulation.SimulateMsgCreateFailure(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateFailure int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgUpdateFailure, &weightMsgUpdateFailure, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateFailure = defaultWeightMsgUpdateFailure
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateFailure,
		contractmanagersimulation.SimulateMsgUpdateFailure(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgDeleteFailure int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgDeleteFailure, &weightMsgDeleteFailure, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteFailure = defaultWeightMsgDeleteFailure
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeleteFailure,
		contractmanagersimulation.SimulateMsgDeleteFailure(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
