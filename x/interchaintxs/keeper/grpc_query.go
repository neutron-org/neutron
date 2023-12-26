package keeper

import (
	"github.com/neutron-org/neutron/v2/x/interchaintxs/types"
)

var _ types.QueryServer = Keeper{}
