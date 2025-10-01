package keeper

import (
	"github.com/neutron-org/neutron/v8/x/interchaintxs/types"
)

var _ types.QueryServer = Keeper{}
