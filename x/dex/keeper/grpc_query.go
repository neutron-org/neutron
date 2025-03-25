package keeper

import (
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

var _ types.QueryServer = Keeper{}
