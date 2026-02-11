package keeper

import (
	"github.com/neutron-org/neutron/v10/x/dynamicfees/types"
)

var _ types.QueryServer = Keeper{}
