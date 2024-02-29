package keeper

import (
	"github.com/neutron-org/neutron/v3/x/dex/types"
)

var _ types.QueryServer = Keeper{}
