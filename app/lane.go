package app

import (
	"cosmossdk.io/math"
	signer_extraction_adapter "github.com/skip-mev/block-sdk/v2/adapters/signer_extraction_adapter"
	blocksdkbase "github.com/skip-mev/block-sdk/v2/block/base"
	base_lane "github.com/skip-mev/block-sdk/v2/lanes/base"
)

const (
	MaxTxsForDefaultLane = 3000 // maximal number of txs that can be stored in this lane at any point in time
)

var MaxBlockspaceForDefaultLane = math.LegacyMustNewDecFromStr("1") // maximal fraction of blockMaxBytes / gas that can be used by this lane at any point in time (90%)

// CreateLanes creates a LaneMempool containing MEV, default lanes (in that order)
func (app *App) CreateLanes() *blocksdkbase.BaseLane {
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

	return baseLane
}
