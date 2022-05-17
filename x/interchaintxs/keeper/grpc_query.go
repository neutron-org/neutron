package keeper

import (
	"github.com/lidofinance/interchain-adapter/x/interchaintxs/types"
)

var _ types.QueryServer = Keeper{}
