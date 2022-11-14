package keeper

import (
	"github.com/neutron-org/neutron/x/adminmodule/types"
)

var _ types.QueryServer = Keeper{}
