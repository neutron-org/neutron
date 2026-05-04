package keeper

import (
	"github.com/neutron-org/neutron/v11/x/globalfee/types"
)

var _ types.QueryServer = Keeper{}
