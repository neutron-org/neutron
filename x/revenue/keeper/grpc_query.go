package keeper

import (
	"github.com/neutron-org/neutron/v5/x/revenue/types"
)

var _ types.QueryServer = Keeper{}
