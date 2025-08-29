package keeper

import (
	"github.com/neutron-org/neutron/v8/x/dynamicfees/types"
)

var _ types.QueryServer = Keeper{}
