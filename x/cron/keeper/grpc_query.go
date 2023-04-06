package keeper

import (
	"github.com/neutron-org/neutron/x/cron/types"
)

var _ types.QueryServer = Keeper{}
