package keeper

import (
	"github.com/neutron-org/neutron/v6/x/interchaintxs/types"
)

var _ types.QueryServer = Keeper{}
