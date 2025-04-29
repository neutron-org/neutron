package app

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/x/staking"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// ExportAppStateAndValidators exports the state of the application for a genesis
// file.
func (app *App) ExportAppStateAndValidators(
	forZeroHeight bool, jailAllowedAddrs, modulesToExport []string,
) (servertypes.ExportedApp, error) {
	// as if they could withdraw from the start of the next block
	ctx := app.NewContext(true)

	// We export at last height + 1, because that's the height at which
	// Tendermint will start InitChain.
	height := app.LastBlockHeight() + 1
	if forZeroHeight {
		height = 0
		err := app.prepForZeroHeightGenesis(ctx, jailAllowedAddrs)
		if err != nil {
			return servertypes.ExportedApp{}, err
		}
	}

	genState, err := app.mm.ExportGenesisForModules(ctx, app.appCodec, modulesToExport)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	appState, err := json.MarshalIndent(genState, "", "  ")
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	validators, err := staking.WriteValidators(ctx, app.StakingKeeper)
	return servertypes.ExportedApp{
		AppState:        appState,
		Validators:      validators,
		Height:          height,
		ConsensusParams: app.BaseApp.GetConsensusParams(ctx),
	}, err
}

// prepare for fresh start at zero height
// NOTE zero height genesis is a temporary feature which will be deprecated
//
//	in favour of export at a block height
func (app *App) prepForZeroHeightGenesis(ctx sdk.Context, _ []string) error {
	/* Just to be safe, assert the invariants on current state. */
	app.CrisisKeeper.AssertInvariants(ctx)

	// set context height to zero
	height := ctx.BlockHeight()
	ctx = ctx.WithBlockHeight(0)

	// reset context height
	ctx = ctx.WithBlockHeight(height)

	/* Handle slashing state. */

	// reset start height on signing infos

	return app.SlashingKeeper.IterateValidatorSigningInfos(
		ctx,
		func(addr sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
			info.StartHeight = 0
			err := app.SlashingKeeper.SetValidatorSigningInfo(ctx, addr, info)
			if err != nil {
				panic(err)
			}
			return false
		},
	)
}
