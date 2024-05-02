package keeper

import (
	"github.com/neutron-org/neutron/v4/x/interchaintxs/types"
)

var _ types.QueryServer = Keeper{}
