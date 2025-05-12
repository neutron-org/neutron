package keeper

import (
	"github.com/neutron-org/neutron/v7/x/interchaintxs/types"
)

var _ types.QueryServer = Keeper{}
