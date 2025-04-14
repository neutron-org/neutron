package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/exp/slices"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/types"
	"github.com/neutron-org/neutron/v6/x/dex/utils"
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
		trancheMaybe := tick.GetLimitOrderTranche()
		if tick.HasToken() && (trancheMaybe == nil || !trancheMaybe.IsExpired(ctx)) {
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

func (k Keeper) GetWhitelistedLPs(ctx sdk.Context) []string {
	return k.GetParams(ctx).WhitelistedLps
}

func (k Keeper) IsWhitelistedLP(ctx sdk.Context, addr sdk.AccAddress) bool {
	whitelistedLPs := k.GetWhitelistedLPs(ctx)
	return slices.Contains(whitelistedLPs, addr.String())
}

func (k Keeper) GetMaxJITsPerBlock(ctx sdk.Context) uint64 {
	return k.GetParams(ctx).MaxJitsPerBlock
}

func (k Keeper) AssertCanPlaceJIT(ctx sdk.Context) error {
	maxJITsAllowed := k.GetMaxJITsPerBlock(ctx)
	JITsInBlock := k.GetJITsInBlockCount(ctx)

	if JITsInBlock == maxJITsAllowed {
		return types.ErrOverJITPerBlockLimit
	}

	return nil
}

func (k Keeper) GetGoodTilPurgeAllowance(ctx sdk.Context) uint64 {
	return k.GetParams(ctx).GoodTilPurgeAllowance
}

func (k Keeper) IsBehindEnemyLines(ctx sdk.Context, tradePairID *types.TradePairID, tickIndex int64) bool {
	oppositeTick, found := k.GetCurrTickIndexTakerToMaker(ctx, tradePairID.Reversed())

	if found && tickIndex*-1 > oppositeTick {
		return true
	}

	return false
}

func (k Keeper) IsPoolBehindEnemyLines(ctx sdk.Context, pairID *types.PairID, tickIndex int64, fee uint64, amount0, amount1 math.Int) bool {
	if amount0.IsPositive() {
		tradePairID0 := types.NewTradePairIDFromMaker(pairID, pairID.Token0)
		tick0 := tickIndex*-1 + utils.MustSafeUint64ToInt64(fee)
		if k.IsBehindEnemyLines(ctx, tradePairID0, tick0) {
			return true
		}
	}

	if amount1.IsPositive() {
		tradePairID1 := types.NewTradePairIDFromMaker(pairID, pairID.Token1)
		tick1 := tickIndex + utils.MustSafeUint64ToInt64(fee)
		if k.IsBehindEnemyLines(ctx, tradePairID1, tick1) {
			return true
		}
	}

	return false
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
	coins sdk.Coins,
) error {
	// transfer tokens to module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, coins); err != nil {
		return err
	}
	// burn tokens
	err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins)

	return err
}
