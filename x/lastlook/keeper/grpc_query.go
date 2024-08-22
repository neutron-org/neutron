package keeper

import (
	"github.com/neutron-org/neutron/v4/x/lastlook/types"
)

var _ types.QueryServer = Keeper{}
