package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/neutron-org/neutron/x/incentives/types"
)

// GetParams returns all of the parameters in the incentive module.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyParams)
	if bz == nil {
		return params
	}

	if err := proto.Unmarshal(bz, &params); err != nil {
		panic(err)
	}

	return params
}

// SetParams sets all of the parameters in the incentive module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := proto.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(types.KeyParams, bz)
	return nil
}
