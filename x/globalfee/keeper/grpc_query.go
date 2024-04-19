package keeper

import (
	"github.com/neutron-org/neutron/v3/x/globalfee/types"
)

var _ types.QueryServer = Keeper{}
