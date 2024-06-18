package keeper

import (
	"github.com/neutron-org/neutron/v4/x/dynamicfees/types"
)

var _ types.QueryServer = Keeper{}
