package keeper

import (
	"github.com/neutron-org/neutron/x/contractmanager/types"
)

var _ types.QueryServer = Keeper{}
