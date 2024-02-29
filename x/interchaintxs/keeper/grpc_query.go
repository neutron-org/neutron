package keeper

import (
	"github.com/neutron-org/neutron/v3/x/interchaintxs/types"
)

var _ types.QueryServer = Keeper{}
