package keeper

import (
	"github.com/neutron-org/neutron/v9/x/interchaintxs/types"
)

var _ types.QueryServer = Keeper{}
