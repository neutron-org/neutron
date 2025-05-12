package keeper

import (
	"github.com/neutron-org/neutron/v7/x/dex/types"
)

var _ types.QueryServer = Keeper{}
