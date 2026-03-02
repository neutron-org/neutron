package keeper

import (
	"github.com/neutron-org/neutron/v10/x/dex/types"
)

var _ types.QueryServer = Keeper{}
