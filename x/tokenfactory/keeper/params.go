package keeper

import (
	"github.com/neutron-org/neutron/v4/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams returns the total set params.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return params
	}

	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets the total set of params.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(types.ParamsKey, bz)
	return nil
}

func (k Keeper) isHookWhitelisted(ctx sdk.Context, denom string, contractAddress sdk.AccAddress) bool {
	contractInfo := k.contractKeeper.GetContractInfo(ctx, contractAddress)
	if contractInfo == nil {
		return false
	}
	codeID := contractInfo.CodeID
	whitelistedHooks := k.GetParams(ctx).WhitelistedHooks
	denomCreator, _, err := types.DeconstructDenom(denom)
	if err != nil {
		return false
	}

	for _, hook := range whitelistedHooks {
		if hook.CodeID == codeID && hook.DenomCreator == denomCreator {
			return true
		}
	}

	return false
}
