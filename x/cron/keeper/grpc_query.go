package keeper

import (
	"github.com/neutron-org/neutron/v3/x/cron/types"
)

var _ types.QueryServer = Keeper{}
