package keeper

import (
	"github.com/lidofinance/interchain-adapter/x/interchainadapter/types"
)

var _ types.QueryServer = Keeper{}
