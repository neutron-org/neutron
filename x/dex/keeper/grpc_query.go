package keeper

import (
	"github.com/neutron-org/neutron/v8/x/dex/types"
)

var _ types.QueryServer = Keeper{}
