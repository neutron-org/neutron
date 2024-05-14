package keeper

import (
	"github.com/neutron-org/neutron/v4/x/cron/types"
)

var _ types.QueryServer = Keeper{}
