package keeper

import (
	"github.com/lidofinance/gaia-wasm-zone/x/interchaintxs/types"
)

var _ types.QueryServer = Keeper{}
