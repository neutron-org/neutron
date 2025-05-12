package keeper

import (
	"github.com/neutron-org/neutron/v7/x/dynamicfees/types"
)

var _ types.QueryServer = Keeper{}
