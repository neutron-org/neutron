package keeper

import (
	"github.com/neutron-org/neutron/v9/x/cron/types"
)

var _ types.QueryServer = Keeper{}
