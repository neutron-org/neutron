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

func (k Keeper) AssertIsHookWhitelisted(ctx sdk.Context, denom string, contractAddress sdk.AccAddress) error {
	contractInfo := k.contractKeeper.GetContractInfo(ctx, contractAddress)
	if contractInfo == nil {
		return types.ErrBeforeSendHookNotWhitelisted.Wrapf("contract with address (%s) does not exist", contractAddress.String())
	}
	codeID := contractInfo.CodeID
	whitelistedHooks := k.GetParams(ctx).WhitelistedHooks
	denomCreator, _, err := types.DeconstructDenom(denom)
	if err != nil {
		return types.ErrBeforeSendHookNotWhitelisted.Wrapf("invalid denom: %s", denom)
	}

	for _, hook := range whitelistedHooks {
		if hook.CodeID == codeID && hook.DenomCreator == denomCreator {
			return nil
		}
	}

	return types.ErrBeforeSendHookNotWhitelisted.Wrapf("no whitelist for contract with codeID (%d) and denomCreator (%s) ", codeID, denomCreator)
}
