package keeper

import (
	"github.com/neutron-org/neutron/v5/x/interchaintxs/types"
)

var _ types.QueryServer = Keeper{}
