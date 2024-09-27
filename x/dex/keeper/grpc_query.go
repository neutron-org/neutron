package keeper

import (
	"github.com/neutron-org/neutron/v5/x/dex/types"
)

var _ types.QueryServer = Keeper{}
