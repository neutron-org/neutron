package keeper

import (
	"github.com/lidofinance/interchain-adapter/x/interchainqueries/types"
)

var _ types.QueryServer = Keeper{}
