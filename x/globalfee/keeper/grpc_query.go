package keeper

import (
	"github.com/neutron-org/neutron/v8/x/globalfee/types"
)

var _ types.QueryServer = Keeper{}
