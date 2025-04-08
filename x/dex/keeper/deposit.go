package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/utils"
	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/types"
	dexutils "github.com/neutron-org/neutron/v6/x/dex/utils"
)

// DepositCore handles core logic for MsgDeposit including bank operations and event emissions
func (k Keeper) DepositCore(
	goCtx context.Context,
	pairID *types.PairID,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
	amounts0 []math.Int,
	amounts1 []math.Int,
	tickIndices []int64,
	fees []uint64,
	options []*types.DepositOptions,
) (amounts0Deposit, amounts1Deposit []math.Int, sharesIssued sdk.Coins, failedDeposits []*types.FailedDeposit, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	amounts0Deposited,
		amounts1Deposited,
		totalInAmount0,
		totalInAmount1,
		sharesIssued,
		events,
		failedDeposits,
		err := k.ExecuteDeposit(ctx, pairID, callerAddr, receiverAddr, amounts0, amounts1, tickIndices, fees, options)
	if err != nil {
		return nil, nil, nil, failedDeposits, err
	}

	ctx.EventManager().EmitEvents(events)

	if totalInAmount0.IsPositive() {
		coin0 := sdk.NewCoin(pairID.Token0, totalInAmount0)
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, callerAddr, types.ModuleName, sdk.Coins{coin0}); err != nil {
			return nil, nil, nil, nil, err
		}
	}

	if totalInAmount1.IsPositive() {
		coin1 := sdk.NewCoin(pairID.Token1, totalInAmount1)
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, callerAddr, types.ModuleName, sdk.Coins{coin1}); err != nil {
			return nil, nil, nil, nil, err
		}
	}

	if err := k.MintShares(ctx, receiverAddr, sharesIssued); err != nil {
		return nil, nil, nil, nil, err
	}

	return amounts0Deposited, amounts1Deposited, sharesIssued, failedDeposits, nil
}

// ExecuteDeposit handles core logic for deposits -- checking and initializing data structures (tick, pair), calculating
// shares based on amount deposited. IT DOES NOT PERFORM ANY BANKING OPERATIONS.
func (k Keeper) ExecuteDeposit(
	ctx sdk.Context,
	pairID *types.PairID,
	callerAddr sdk.AccAddress,
	receiverAddr sdk.AccAddress,
	amounts0 []math.Int,
	amounts1 []math.Int,
	tickIndices []int64,
	fees []uint64,
	options []*types.DepositOptions) (
	amounts0Deposited, amounts1Deposited []math.Int,
	totalInAmount0, totalInAmount1 math.Int,
	sharesIssued sdk.Coins,
	events sdk.Events,
	failedDeposits []*types.FailedDeposit,
	err error,
) {
	totalInAmount0 = math.ZeroInt()
	totalInAmount1 = math.ZeroInt()
	amounts0Deposited = make([]math.Int, len(amounts0))
	amounts1Deposited = make([]math.Int, len(amounts1))
	sharesIssued = sdk.Coins{}

	for i := 0; i < len(amounts0); i++ {
		amounts0Deposited[i] = math.ZeroInt()
		amounts1Deposited[i] = math.ZeroInt()
	}
	isWhitelistedLP := k.IsWhitelistedLP(ctx, callerAddr)

	for i, depositAmount0 := range amounts0 {
		depositAmount1 := amounts1[i]
		tickIndex := tickIndices[i]
		fee := fees[i]
		option := options[i]
		inAmount0, inAmount1 := depositAmount0, depositAmount1
		if option == nil {
			option = &types.DepositOptions{}
		}
		autoswap := !option.DisableAutoswap

		// Enforce deposits only at valid fee tiers. This does not apply to whitelistedLPs
		if !isWhitelistedLP {
			if err := k.ValidateFee(ctx, fee); err != nil {
				return nil, nil, math.ZeroInt(), math.ZeroInt(), nil, nil, nil, err
			}
		}
		if option.SwapOnDeposit {
			inAmount0, inAmount1, depositAmount0, depositAmount1, err = k.SwapOnDeposit(ctx, pairID, tickIndex, fee, depositAmount0, depositAmount1, option.SwapOnDepositSlopToleranceBps)
			if err != nil {
				return nil, nil, math.ZeroInt(), math.ZeroInt(), nil, nil, nil, err
			}
		}

		// This check is redundant when using SwapOnDepsit. But we leave it as an extra check.
		if k.IsPoolBehindEnemyLines(ctx, pairID, tickIndex, fee, depositAmount0, depositAmount1) {
			err = sdkerrors.Wrapf(types.ErrDepositBehindEnemyLines,
				"deposit failed at tick %d fee %d", tickIndex, fee)
			if option.FailTxOnBel {
				return nil, nil, math.ZeroInt(), math.ZeroInt(), nil, nil, nil, err
			}
			failedDeposits = append(failedDeposits, &types.FailedDeposit{DepositIdx: uint64(i), Error: err.Error()}) //nolint:gosec
			continue
		}

		pool, err := k.GetOrInitPool(
			ctx,
			pairID,
			tickIndex,
			fee,
		)
		if err != nil {
			return nil, nil, math.ZeroInt(), math.ZeroInt(), nil, nil, nil, err
		}

		existingShares := k.bankKeeper.GetSupply(ctx, pool.GetPoolDenom()).Amount

		depositAmount0, depositAmount1, outShares := pool.Deposit(depositAmount0, depositAmount1, existingShares, autoswap)
		if option.DisableAutoswap {
			// If autoswap is disabled inAmount might change.
			// SwapOnDeposit cannot be used without autoswap, so nothing is affected here.
			inAmount0, inAmount1 = depositAmount0, depositAmount1
		}

		// Save updates to both sides of the pool
		k.UpdatePool(ctx, pool)

		if inAmount0.IsZero() && inAmount1.IsZero() {
			return nil, nil, math.ZeroInt(), math.ZeroInt(), nil, nil, nil, types.ErrZeroTrueDeposit
		}

		if outShares.IsZero() {
			return nil, nil, math.ZeroInt(), math.ZeroInt(), nil, nil, nil, types.ErrDepositShareUnderflow
		}

		sharesIssued = append(sharesIssued, outShares)

		amounts0Deposited[i] = depositAmount0
		amounts1Deposited[i] = depositAmount1
		totalInAmount0 = totalInAmount0.Add(inAmount0)
		totalInAmount1 = totalInAmount1.Add(inAmount1)

		depositEvent := types.CreateDepositEvent(
			callerAddr,
			receiverAddr,
			pairID.Token0,
			pairID.Token1,
			tickIndex,
			fee,
			depositAmount0,
			depositAmount1,
			inAmount0,
			inAmount1,
			outShares.Amount,
		)
		events = append(events, depositEvent)
	}

	// At this point shares issued is not sorted and may have duplicates
	// we must sanitize to convert it to a valid set of coins
	sharesIssued = utils.SanitizeCoins(sharesIssued)
	return amounts0Deposited, amounts1Deposited, totalInAmount0, totalInAmount1, sharesIssued, events, failedDeposits, nil
}

func (k Keeper) PerformSwapOnDepositSwap(ctx sdk.Context, tradePairID *types.TradePairID, amountIn math.Int, limitPrice math_utils.PrecDec, slopToleranceBPs uint64) (inAmount, outAmount math.Int, orderFilled bool, err error) {
	swapTokenIn, swapTokenOut, orderFilled, err := k.Swap(ctx, tradePairID, amountIn, nil, &limitPrice)
	if err != nil {
		return math.ZeroInt(), math.ZeroInt(), false, err
	}

	if err := CheckSwapOnDepositSlopTolerance(swapTokenIn.Amount, swapTokenOut.Amount, limitPrice, slopToleranceBPs); err != nil {
		return math.ZeroInt(), math.ZeroInt(), false, err
	}

	return swapTokenIn.Amount, swapTokenOut.Amount, orderFilled, nil
}

func (k Keeper) SwapOnDeposit(
	ctx sdk.Context,
	pairID *types.PairID,
	tickIndex int64,
	fee uint64,
	amount0, amount1 math.Int,
	slopToleranceBPs uint64,
) (inAmount0, inAmount1, depositAmount0, depositAmount1 math.Int, err error) {
	feeInt64 := dexutils.MustSafeUint64ToInt64(fee)
	inAmount0, inAmount1 = amount0, amount1
	depositAmount0, depositAmount1 = inAmount0, inAmount1
	swappedToken0 := false
	if amount0.IsPositive() {
		// Use Amount0 to swap any Token1 ticks < (-depositTick0 -1 )
		depositTickToken0 := -tickIndex + feeInt64
		// subtract 1 from limit price because we can have double-sided liquidity at the deposit tick
		// we don't want to swap through this liquidity
		limitPrice0 := types.MustCalcPrice(-depositTickToken0 - 1)
		tradePairID := types.MustNewTradePairID(pairID.Token0, pairID.Token1)

		swapAmountIn0, swapAmountOut1, orderFilled, err := k.PerformSwapOnDepositSwap(ctx, tradePairID, amount0, limitPrice0, slopToleranceBPs)
		if err != nil {
			return math.ZeroInt(), math.ZeroInt(), math.ZeroInt(), math.ZeroInt(), err
		}

		if swapAmountIn0.IsPositive() {
			if orderFilled {
				// due to monotonic rounding we may not be able to swap all of TokenIn
				// but we can't deposit the remainder because it's still behind enemy lines so we don't use it
				inAmount0 = swapAmountIn0
			} // else inAmount = amountIn  we have swapped through all opposing BEL liquidity, so we can safely deposit the full amount

			depositAmount0 = inAmount0.Sub(swapAmountIn0)
			depositAmount1 = amount1.Add(swapAmountOut1)
			inAmount1 = amount1
			swappedToken0 = true
		}
	}

	if amount1.IsPositive() {
		// Use amount1 to swap any Token0 ticks < (-depositTick1 - 1)
		depositTickToken1 := tickIndex + feeInt64
		limitPrice1 := types.MustCalcPrice(-depositTickToken1 - 1)
		tradePairID := types.MustNewTradePairID(pairID.Token1, pairID.Token0)

		swapAmountIn1, swapAmountOut0, orderFilled, err := k.PerformSwapOnDepositSwap(ctx, tradePairID, amount1, limitPrice1, slopToleranceBPs)
		if err != nil {
			return math.ZeroInt(), math.ZeroInt(), math.ZeroInt(), math.ZeroInt(), err
		}

		if swapAmountIn1.IsPositive() {
			if swappedToken0 {
				// This should be impossible, but leaving this as an extra precaution
				return math.ZeroInt(), math.ZeroInt(), math.ZeroInt(), math.ZeroInt(), types.ErrDoubleSidedSwapOnDeposit
			}

			if orderFilled { // see note above on monotonic rounding logic
				inAmount1 = swapAmountIn1
			} // else inAmount1 = amount1
			depositAmount0 = amount0.Add(swapAmountOut0)
			depositAmount1 = inAmount1.Sub(swapAmountIn1)
			inAmount0 = amount0

		}

	}

	return inAmount0, inAmount1, depositAmount0, depositAmount1, nil
}

func CheckSwapOnDepositSlopTolerance(swapAmountIn, swapAmountOut math.Int, limitPrice math_utils.PrecDec, slopToleranceBPs uint64) error {
	if swapAmountIn.IsPositive() {
		trueTakerPrice := math_utils.NewPrecDecFromInt(swapAmountIn).QuoInt(swapAmountOut)
		// slopToleranceBPs has already been validated so no risk of overflow
		slopToleranceInt64 := dexutils.MustSafeUint64ToInt64(slopToleranceBPs)
		slopToleranceDec := math_utils.NewPrecDec(slopToleranceInt64).Quo(math_utils.NewPrecDecFromInt(math.NewInt(10000)))
		maxAllowedTakerPrice := limitPrice.Mul(math_utils.OnePrecDec().Add(slopToleranceDec))
		if trueTakerPrice.GTE(maxAllowedTakerPrice) {
			return types.ErrSwapOnDepositSlopToleranceNotSatisfied
		}
	}
	return nil
}
