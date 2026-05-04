package keeper

import (
	"github.com/neutron-org/neutron/v11/x/interchaintxs/types"
)

var _ types.QueryServer = Keeper{}
