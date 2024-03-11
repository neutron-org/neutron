package keeper

import (
	"github.com/neutron-org/neutron/v3/x/contractmanager/types"
)

var _ types.QueryServer = Keeper{}
