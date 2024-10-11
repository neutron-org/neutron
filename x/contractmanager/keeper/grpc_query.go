package keeper

import (
	"github.com/neutron-org/neutron/v5/x/contractmanager/types"
)

var _ types.QueryServer = Keeper{}
