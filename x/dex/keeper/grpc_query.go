package keeper

import (
	"github.com/neutron-org/neutron/x/dex/types"
)

var _ types.QueryServer = Keeper{}
