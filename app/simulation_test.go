package app_test

// TODO: enable test
// at the moment latest ibc-go release requires outdated app interface in the `simapp.SimulationOperations` method

//
//import (
//	"github.com/cosmos/cosmos-sdk/server/types"
//	"github.com/cosmos/ibc-go/v7/testing/simapp"
//	"os"
//	"testing"
//
//	abci "github.com/cometbft/cometbft/abci/types"
//	"github.com/cosmos/cosmos-sdk/baseapp"
//	"github.com/cosmos/cosmos-sdk/codec"
//	sdk "github.com/cosmos/cosmos-sdk/types"
//	"github.com/cosmos/cosmos-sdk/types/module"
//	simulationtypes "github.com/cosmos/cosmos-sdk/types/simulation"
//	"github.com/cosmos/cosmos-sdk/x/simulation"
//	"github.com/stretchr/testify/require"
//
//	"github.com/neutron-org/neutron/app"
//)
//
//func init() {
//	simapp.GetSimulatorFlags()
//}
//
//type SimApp interface {
//	GetBaseApp() *baseapp.BaseApp
//	AppCodec() codec.Codec
//	SimulationManager() *module.SimulationManager
//	ModuleAccountAddrs() map[string]bool
//	Name() string
//	LegacyAmino() *codec.LegacyAmino
//	BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock
//	EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock
//	InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain
//
//	// Loads the app at a given height.
//	LoadHeight(height int64) error
//
//	// Exports the state of the application for a genesis file.
//	ExportAppStateAndValidators(
//		forZeroHeight bool, jailAllowedAddrs []string,
//	) (types.ExportedApp, error)
//}
//
//// BenchmarkSimulation run the chain simulation
//// Running using starport command:
//// `starport chain simulate -v --numBlocks 200 --blockSize 50`
//// Running as go benchmark test:
//// `go test -benchmem -run=^$ -bench ^BenchmarkSimulation ./app -NumBlocks=200 -BlockSize 50 -Commit=true -Verbose=true -Enabled=true`
//func BenchmarkSimulation(b *testing.B) {
//	simapp.FlagEnabledValue = true
//	simapp.FlagCommitValue = true
//
//	config, db, dir, logger, _, err := simapp.SetupSimulation("goleveldb-app-sim", "Simulation")
//	require.NoError(b, err, "simulation setup failed")
//
//	b.Cleanup(func() {
//		db.Close()
//		err = os.RemoveAll(dir)
//		require.NoError(b, err)
//	})
//
//	encoding := app.MakeEncodingConfig()
//
//	app := app.New(
//		logger,
//		db,
//		nil,
//		true,
//		map[int64]bool{},
//		app.DefaultNodeHome,
//		0,
//		encoding,
//		app.GetEnabledProposals(),
//		simapp.EmptyAppOptions{},
//		nil,
//	)
//
//	simApp := app
//	// require.True(b, ok, "can't use simapp")
//
//	// Run randomized simulations
//	_, simParams, simErr := simulation.SimulateFromSeed(
//		b,
//		os.Stdout,
//		simApp.GetBaseApp(),
//		simapp.AppStateFn(simApp.AppCodec(), simApp.SimulationManager()),
//		simulationtypes.RandomAccounts,
//		simapp.SimulationOperations(app, simApp.AppCodec(), config),
//		simApp.ModuleAccountAddrs(),
//		config,
//		simApp.AppCodec(),
//	)
//
//	// export state and simParams before the simulation error is checked
//	err = simapp.CheckExportSimulation(app, config, simParams)
//	require.NoError(b, err)
//	require.NoError(b, simErr)
//
//	if config.Commit {
//		simapp.PrintStats(db)
//	}
//}
