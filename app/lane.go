package app

import (
	"cosmossdk.io/math"
	signer_extraction_adapter "github.com/skip-mev/block-sdk/adapters/signer_extraction_adapter"
	blocksdkbase "github.com/skip-mev/block-sdk/block/base"
	base_lane "github.com/skip-mev/block-sdk/lanes/base"
	"github.com/skip-mev/block-sdk/lanes/mev"
	mev_lane "github.com/skip-mev/block-sdk/lanes/mev"
)

const (
	MaxTxsForDefaultLane = 3000
	MaxTxsForMEVLane = 500
)

var (
	MaxBlockspaceForDefaultLane = math.LegacyMustNewDecFromStr("0.9")
	MaxBlockspaceForMEVLane = math.LegacyMustNewDecFromStr("0.1")
)

// CreateLanes creates a LaneMempool containing MEV, free lanes (in that order)
func (app *App) CreateLanes() (*mev.MEVLane, *blocksdkbase.BaseLane) {
	// initialize lanes
	basecfg := blocksdkbase.LaneConfig{
		Logger:          app.Logger(),
		TxDecoder:       app.GetTxConfig().TxDecoder(),
		TxEncoder:       app.GetTxConfig().TxEncoder(),
		SignerExtractor: signer_extraction_adapter.NewDefaultAdapter(),
		MaxBlockSpace:   MaxBlockspaceForDefaultLane,
		MaxTxs:          MaxTxsForDefaultLane,
	}

	baseLane := base_lane.NewDefaultLane(basecfg, blocksdkbase.DefaultMatchHandler())

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
	app.MEVLane = mevLane

	return mevLane, baseLane
}
