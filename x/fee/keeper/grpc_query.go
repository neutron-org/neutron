package keeper

import (
	"github.com/neutron-org/neutron/x/fee/types"
)

var _ types.QueryServer = Keeper{}
