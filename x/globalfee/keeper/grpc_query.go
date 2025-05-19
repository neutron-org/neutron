package keeper

import (
	"github.com/neutron-org/neutron/v7/x/globalfee/types"
)

var _ types.QueryServer = Keeper{}
