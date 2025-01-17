package app

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signer_extraction_adapter "github.com/skip-mev/block-sdk/v2/adapters/signer_extraction_adapter"
	blocksdkbase "github.com/skip-mev/block-sdk/v2/block/base"
	base_lane "github.com/skip-mev/block-sdk/v2/lanes/base"
	free_lane "github.com/skip-mev/block-sdk/v2/lanes/free"
)

const (
	MaxTxsForDefaultLane = 3000 // maximal number of txs that can be stored in this lane at any point in time
)

var (
	MaxBlockSpaceForFreeLane    = math.LegacyMustNewDecFromStr("0.1") // (10%)
	MaxBlockspaceForDefaultLane = math.LegacyMustNewDecFromStr("0.9") // maximal fraction of blockMaxBytes / gas that can be used by this lane at any point in time (90%)
)

// CreateLanes creates a LaneMempool containing MEV, default lanes (in that order)
func (app *App) CreateLanes() (*blocksdkbase.BaseLane, *blocksdkbase.BaseLane) {
	// initialize the free lane
	freeConfig := blocksdkbase.LaneConfig{
		Logger:          app.Logger(),
		TxEncoder:       app.GetTxConfig().TxEncoder(),
		TxDecoder:       app.GetTxConfig().TxDecoder(),
		SignerExtractor: signer_extraction_adapter.NewDefaultAdapter(),
		MaxBlockSpace:   MaxBlockSpaceForFreeLane,
		MaxTxs:          0,
	}

	freeLane := free_lane.NewFreeLane(freeConfig, blocksdkbase.NewDefaultTxPriority(), app.freeLaneMatchHandler())

	// initialize the default lane
	basecfg := blocksdkbase.LaneConfig{
		Logger:          app.Logger(),
		TxDecoder:       app.GetTxConfig().TxDecoder(),
		TxEncoder:       app.GetTxConfig().TxEncoder(),
		SignerExtractor: signer_extraction_adapter.NewDefaultAdapter(),
		MaxBlockSpace:   MaxBlockspaceForDefaultLane,
		MaxTxs:          MaxTxsForDefaultLane,
	}

	// BaseLane (DefaultLane) is intended to hold all txs that are not matched by any lanes ordered before this
	// lane.
	baseLane := base_lane.NewDefaultLane(basecfg, blocksdkbase.DefaultMatchHandler())
	baseLane.LaneMempool = blocksdkbase.NewMempool(
		blocksdkbase.NewDefaultTxPriority(),
		basecfg.SignerExtractor,
		basecfg.MaxTxs,
	)

	return freeLane, baseLane
}

func (app *App) freeLaneMatchHandler() blocksdkbase.MatchHandler {
	return func(ctx sdk.Context, tx sdk.Tx) bool {
		feeTx, ok := tx.(sdk.FeeTx)
		if !ok {
			return false
		}

		acc := app.AccountKeeper.GetAccount(ctx, feeTx.FeePayer())

		freeLaneParams := app.FreeLaneKeeper.GetParams(ctx)

		return acc.GetSequence() < freeLaneParams.SequenceNumber
	}
}
