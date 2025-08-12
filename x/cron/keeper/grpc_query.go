package keeper

import (
	"github.com/neutron-org/neutron/v8/x/cron/types"
)

var _ types.QueryServer = Keeper{}
