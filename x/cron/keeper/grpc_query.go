package keeper

import (
	"github.com/neutron-org/neutron/v6/x/cron/types"
)

var _ types.QueryServer = Keeper{}
