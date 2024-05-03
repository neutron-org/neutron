package keeper

import (
	"github.com/neutron-org/neutron/v4/x/globalfee/types"
)

var _ types.QueryServer = Keeper{}
