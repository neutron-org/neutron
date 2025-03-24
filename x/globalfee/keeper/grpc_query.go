package keeper

import (
	"github.com/neutron-org/neutron/v6/x/globalfee/types"
)

var _ types.QueryServer = Keeper{}
