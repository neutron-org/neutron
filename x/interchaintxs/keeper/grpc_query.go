package keeper

import (
	"github.com/neutron-org/neutron/x/interchaintxs/types"
)

var _ types.QueryServer = Keeper{}
