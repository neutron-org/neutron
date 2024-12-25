package keeper

import (
	"github.com/neutron-org/neutron/v5/x/harpoon/types"
)

var _ types.QueryServer = Keeper{}
