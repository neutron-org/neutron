package keeper

import (
	"github.com/neutron-org/neutron/v9/x/globalfee/types"
)

var _ types.QueryServer = Keeper{}
