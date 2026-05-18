package keeper

import (
	"github.com/neutron-org/neutron/v11/x/dynamicfees/types"
)

var _ types.QueryServer = Keeper{}
