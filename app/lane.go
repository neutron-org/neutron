package app

import (
	"cosmossdk.io/math"
	signer_extraction_adapter "github.com/skip-mev/block-sdk/adapters/signer_extraction_adapter"
	blocksdkbase "github.com/skip-mev/block-sdk/block/base"
	base_lane "github.com/skip-mev/block-sdk/lanes/base"
	mev_lane "github.com/skip-mev/block-sdk/lanes/mev"
)

const (
	MaxTxsForDefaultLane = 3000 // maximal number of txs that can be stored in this lane at any point in time
	MaxTxsForMEVLane     = 500  // ditto
)

var (
	MaxBlockspaceForDefaultLane = math.LegacyMustNewDecFromStr("0.9") // maximal fraction of blockMaxBytes / gas that can be used by this lane at any point in time (90%)
	MaxBlockspaceForMEVLane     = math.LegacyMustNewDecFromStr("0.1") // ditto (10%)
)

// CreateLanes creates a LaneMempool containing MEV, default lanes (in that order)
func (app *App) CreateLanes() (*mev_lane.MEVLane, *blocksdkbase.BaseLane) {
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

	// initialize the MEV lane, this lane is intended to hold all bid txs
	factory := mev_lane.NewDefaultAuctionFactory(app.GetTxConfig().TxDecoder(), signer_extraction_adapter.NewDefaultAdapter())

	mevcfg := blocksdkbase.LaneConfig{
		Logger:          app.Logger(),
		TxDecoder:       app.GetTxConfig().TxDecoder(),
		TxEncoder:       app.GetTxConfig().TxEncoder(),
		SignerExtractor: signer_extraction_adapter.NewDefaultAdapter(),
		MaxBlockSpace:   MaxBlockspaceForMEVLane,
		MaxTxs:          MaxTxsForMEVLane,
	}
	mevLane := mev_lane.NewMEVLane(
		mevcfg,
		factory,
		factory.MatchHandler(),
	)

	// set the MEVLane used in the mev-ante-handler on the application
	app.MEVLane = mevLane

	return mevLane, baseLane
}
