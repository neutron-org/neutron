package keeper

import (
	"context"

	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
)

// TODO: better errors msgs

// SetParams sets the x/revenue module parameters.
func (k *Keeper) SetParams(ctx context.Context, params revenuetypes.Params) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {

		return err
	}
	return store.Set(revenuetypes.ParamsKey, bz)
}

// GetParams gets the x/revenue module parameters.
func (k *Keeper) GetParams(ctx context.Context) (params revenuetypes.Params, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(revenuetypes.ParamsKey)
	if err != nil {
		return params, err
	}

	if bz == nil {
		return params, nil
	}

	err = k.cdc.Unmarshal(bz, &params)
	return params, err
}
