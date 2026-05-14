package keeper

import (
	"github.com/neutron-org/neutron/v11/x/dex/types"
)

var _ types.QueryServer = Keeper{}
