package keeper

import (
	"github.com/neutron-org/neutron/v2/x/contractmanager/types"
)

var _ types.QueryServer = Keeper{}
