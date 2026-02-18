package keeper

import (
	"github.com/neutron-org/neutron/v10/x/interchaintxs/types"
)

var _ types.QueryServer = Keeper{}
