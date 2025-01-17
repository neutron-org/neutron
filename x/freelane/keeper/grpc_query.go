package keeper

import (
	"github.com/neutron-org/neutron/v5/x/freelane/types"
)

var _ types.QueryServer = Keeper{}
