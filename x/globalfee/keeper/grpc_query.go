package keeper

import (
	"github.com/neutron-org/neutron/v5/x/globalfee/types"
)

var _ types.QueryServer = Keeper{}
