package keeper

import (
	"github.com/neutron-org/neutron/v6/x/dynamicfees/types"
)

var _ types.QueryServer = Keeper{}
