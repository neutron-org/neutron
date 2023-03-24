package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/x/cron/types"
)

// GetParams get all parameters as types.Params
func (k *Keeper) GetParams(_ sdk.Context) types.Params {
	return types.NewParams()
}

// SetParams set the params
func (k *Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
