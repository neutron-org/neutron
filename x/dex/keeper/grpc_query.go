package keeper

import (
	"github.com/neutron-org/neutron/v2/x/dex/types"
)

var _ types.QueryServer = Keeper{}
