package keeper

import (
	"github.com/neutron-org/neutron/v5/x/dynamicfees/types"
)

var _ types.QueryServer = Keeper{}
