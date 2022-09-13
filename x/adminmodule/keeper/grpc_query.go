package keeper

import (
	"github.com/cosmos/admin-module/x/adminmodule/types"
)

var _ types.QueryServer = Keeper{}
