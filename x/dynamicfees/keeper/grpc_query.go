package keeper

import (
	"github.com/neutron-org/neutron/v9/x/dynamicfees/types"
)

var _ types.QueryServer = Keeper{}
