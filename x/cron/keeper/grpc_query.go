package keeper

import (
	"github.com/neutron-org/neutron/v11/x/cron/types"
)

var _ types.QueryServer = Keeper{}
