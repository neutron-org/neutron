package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/slices"

	math_utils "github.com/neutron-org/neutron/v3/utils/math"
	"github.com/neutron-org/neutron/v3/x/dex/types"
)

///////////////////////////////////////////////////////////////////////////////
//                          STATE CALCULATIONS                               //
///////////////////////////////////////////////////////////////////////////////

func (k Keeper) GetCurrPrice(ctx sdk.Context, tradePairID *types.TradePairID) (math_utils.PrecDec, bool) {
	liq := k.GetCurrLiq(ctx, tradePairID)
	if liq != nil {
		return liq.Price(), true
	}
	return math_utils.ZeroPrecDec(), false
}

// Returns a takerToMaker tick index
func (k Keeper) GetCurrTickIndexTakerToMaker(
	ctx sdk.Context,
	tradePairID *types.TradePairID,
) (int64, bool) {
	liq := k.GetCurrLiq(ctx, tradePairID)
	if liq != nil {
		return liq.TickIndex(), true
	}
	return 0, false
}

// Returns a takerToMaker tick index
func (k Keeper) GetCurrTickIndexTakerToMakerNormalized(
	ctx sdk.Context,
	tradePairID *types.TradePairID,
) (int64, bool) {
	tickIndexTakerToMaker, found := k.GetCurrTickIndexTakerToMaker(ctx, tradePairID)
	if found {
		tickIndexTakerToMakerNormalized := tradePairID.TickIndexNormalized(tickIndexTakerToMaker)
		return tickIndexTakerToMakerNormalized, true
	}

	return 0, false
}

func (k Keeper) GetCurrLiq(ctx sdk.Context, tradePairID *types.TradePairID) *types.TickLiquidity {
	ti := k.NewTickIterator(ctx, tradePairID)
	defer ti.Close()
	for ; ti.Valid(); ti.Next() {
		tick := ti.Value()
		if tick.HasToken() {
			return &tick
		}
	}

	return nil
}

func (k Keeper) GetValidFees(ctx sdk.Context) []uint64 {
	return k.GetParams(ctx).FeeTiers
}

func (k Keeper) ValidateFee(ctx sdk.Context, fee uint64) error {
	validFees := k.GetValidFees(ctx)
	if !slices.Contains(validFees, fee) {
		return sdkerrors.Wrapf(types.ErrInvalidFee, "%d", validFees)
	}

	return nil
}

func (k Keeper) IsBehindEnemyLines(ctx sdk.Context, tradePairID *types.TradePairID, tickIndex int64) bool {
	oppositeTick, found := k.GetCurrTickIndexTakerToMaker(ctx, tradePairID.Reversed())

	if found && tickIndex*-1 > oppositeTick {
		return true
	}

	return false
}

func (k Keeper) SwapPoolBehindEnemyLines(ctx sdk.Context, pool *types.Pool, amount0, amount1 math.Int) (newAmount0, newAmount1 math.Int, err error) {

	token0IsBEL := k.IsBehindEnemyLines(ctx, pool.LowerTick0.Key.TradePairId, pool.LowerTick0.Key.TickIndexTakerToMaker)
	token1IsBEL := k.IsBehindEnemyLines(ctx, pool.UpperTick1.Key.TradePairId, pool.UpperTick1.Key.TickIndexTakerToMaker)

	if token0IsBEL && token1IsBEL {
		return math.ZeroInt(), math.ZeroInt(), types.ErrDepositBothSidesBEL
	} else if token0IsBEL && amount0.IsPositive() {
		coin0, additional1Coin, _, err := k.Swap(ctx, pool.LowerTick0.Key.TradePairId, amount0, nil, &pool.LowerTick0.PriceTakerToMaker)
		if err != nil {
			return math.ZeroInt(), math.ZeroInt(), err
		}

		return coin0.Amount, amount1.Add(additional1Coin.Amount), nil

	} else if token1IsBEL && amount1.IsPositive() {
		coin1, additional0Coin, _, err := k.Swap(ctx, pool.UpperTick1.Key.TradePairId, amount1, nil, &pool.UpperTick1.PriceTakerToMaker)
		if err != nil {
			return math.ZeroInt(), math.ZeroInt(), err
		}

		return amount0.Add(additional0Coin.Amount), coin1.Amount, nil

	}

	return amount0, amount1, nil
}

///////////////////////////////////////////////////////////////////////////////
//                            TOKENIZER UTILS                                //
///////////////////////////////////////////////////////////////////////////////

func (k Keeper) MintShares(ctx sdk.Context, addr sdk.AccAddress, sharesCoins sdk.Coins) error {
	// mint share tokens
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sharesCoins); err != nil {
		return err
	}
	// transfer them to addr
	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sharesCoins)

	return err
}

func (k Keeper) BurnShares(
	ctx sdk.Context,
	addr sdk.AccAddress,
	amount math.Int,
	sharesID string,
) error {
	sharesCoins := sdk.Coins{sdk.NewCoin(sharesID, amount)}
	// transfer tokens to module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, sharesCoins); err != nil {
		return err
	}
	// burn tokens
	err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sharesCoins)

	return err
}
