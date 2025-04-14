package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"

	appparams "github.com/neutron-org/neutron/v6/app/params"

	"github.com/neutron-org/neutron/v6/x/dynamicfees/types"
)

var _ feemarkettypes.DenomResolver = Keeper{}

// ConvertToDenom converts NTRN deccoin into the equivalent amount of the token denominated in denom.
func (k Keeper) ConvertToDenom(ctx sdk.Context, fromCoin sdk.DecCoin, toDenom string) (sdk.DecCoin, error) {
	params := k.GetParams(ctx)
	for _, c := range params.NtrnPrices {
		if c.Denom == toDenom && fromCoin.Denom == appparams.DefaultDenom {
			// converts NTRN into the denom
			return sdk.NewDecCoinFromDec(toDenom, fromCoin.Amount.Quo(c.Amount)), nil
		} else if toDenom == appparams.DefaultDenom && fromCoin.Denom == c.Denom {
			// converts the denom into NTRN
			return sdk.NewDecCoinFromDec(appparams.DefaultDenom, fromCoin.Amount.Mul(c.Amount)), nil
		}
	}
	return sdk.DecCoin{}, types.ErrUnknownDenom
}

func (k Keeper) ExtraDenoms(ctx sdk.Context) ([]string, error) {
	params := k.GetParams(ctx)
	denoms := make([]string, 0, params.NtrnPrices.Len())
	for _, coin := range params.NtrnPrices {
		denoms = append(denoms, coin.Denom)
	}
	return denoms, nil
}
