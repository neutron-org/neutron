package keeper

import (
	"github.com/neutron-org/neutron/v10/x/globalfee/types"
)

var _ types.QueryServer = Keeper{}
