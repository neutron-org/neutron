package keeper

import (
	"github.com/neutron-org/neutron/v4/x/dex/types"
)

var _ types.QueryServer = Keeper{}
