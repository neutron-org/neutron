package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v7/utils"
	math_utils "github.com/neutron-org/neutron/v7/utils/math"
	"github.com/neutron-org/neutron/v7/x/dex/types"
	dexutils "github.com/neutron-org/neutron/v7/x/dex/utils"
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
) (amounts0Deposit, amounts1Deposit []math_utils.PrecDec, sharesIssued sdk.Coins, failedDeposits []*types.FailedDeposit, err error) {
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

	coin0 := types.NewPrecDecCoin(pairID.Token0, totalInAmount0)
	coin1 := types.NewPrecDecCoin(pairID.Token1, totalInAmount1)

	// NewPrecDecCoins will remove zero amounts
	coinsToSend := types.NewPrecDecCoins(coin0, coin1)

	if !coinsToSend.Empty() {
		if err := k.FractionalBanker.SendFractionalCoinsFromAccountToDex(ctx, callerAddr, coinsToSend); err != nil {
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
	amounts0Deposited, amounts1Deposited []math_utils.PrecDec,
	totalInAmount0, totalInAmount1 math_utils.PrecDec,
	sharesIssued sdk.Coins,
	events sdk.Events,
	failedDeposits []*types.FailedDeposit,
	err error,
) {
	totalInAmount0 = math_utils.ZeroPrecDec()
	totalInAmount1 = math_utils.ZeroPrecDec()
	amounts0Deposited = make([]math_utils.PrecDec, len(amounts0))
	amounts1Deposited = make([]math_utils.PrecDec, len(amounts1))
	sharesIssued = sdk.Coins{}

	for i := 0; i < len(amounts0); i++ {
		amounts0Deposited[i] = math_utils.ZeroPrecDec()
		amounts1Deposited[i] = math_utils.ZeroPrecDec()
	}
	isWhitelistedLP := k.IsWhitelistedLP(ctx, callerAddr)

	for i, depositAmount0Int := range amounts0 {
		depositAmount0 := math_utils.NewPrecDecFromInt(depositAmount0Int)
		depositAmount1 := math_utils.NewPrecDecFromInt(amounts1[i])
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
				return nil, nil, math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), nil, nil, nil, err
			}
		}

		if option.SwapOnDeposit {
			inAmount0, inAmount1, depositAmount0, depositAmount1, err = k.SwapOnDeposit(ctx, pairID, tickIndex, fee, depositAmount0, depositAmount1)
			if err != nil {
				return nil, nil, math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), nil, nil, nil, err
			}
		}

		// This check is redundant when using SwapOnDepsit. But we leave it as an extra check.
		if k.IsPoolBehindEnemyLines(ctx, pairID, tickIndex, fee, depositAmount0, depositAmount1) {
			err = sdkerrors.Wrapf(types.ErrDepositBehindEnemyLines,
				"deposit failed at tick %d fee %d", tickIndex, fee)
			if option.FailTxOnBel {
				return nil, nil, math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), nil, nil, nil, err
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
			return nil, nil, math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), nil, nil, nil, err
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
			return nil, nil, math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), nil, nil, nil, types.ErrZeroTrueDeposit
		}

		if outShares.IsZero() {
			return nil, nil, math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), nil, nil, nil, types.ErrDepositShareUnderflow
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
			pool.Id,
			outShares.Amount,
		)
		events = append(events, depositEvent)
	}

	// At this point shares issued is not sorted and may have duplicates
	// we must sanitize to convert it to a valid set of coins
	sharesIssued = utils.SanitizeCoins(sharesIssued)
	return amounts0Deposited, amounts1Deposited, totalInAmount0, totalInAmount1, sharesIssued, events, failedDeposits, nil
}

func (k Keeper) PerformSwapOnDepositSwap(ctx sdk.Context, tradePairID *types.TradePairID, amountIn, limitPrice math_utils.PrecDec) (inAmount, outAmount math_utils.PrecDec, orderFilled bool, err error) {
	swapTokenIn, swapTokenOut, orderFilled, err := k.Swap(ctx, tradePairID, amountIn, nil, &limitPrice)
	if err != nil {
		return math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), false, err
	}

	return swapTokenIn.Amount, swapTokenOut.Amount, orderFilled, nil
}

func (k Keeper) SwapOnDeposit(
	ctx sdk.Context,
	pairID *types.PairID,
	tickIndex int64,
	fee uint64,
	amount0, amount1 math_utils.PrecDec,
) (inAmount0, inAmount1, depositAmount0, depositAmount1 math_utils.PrecDec, err error) {
	feeInt64 := dexutils.MustSafeUint64ToInt64(fee)
	inAmount0, inAmount1 = amount0, amount1
	depositAmount0, depositAmount1 = amount0, amount1
	swappedToken0 := false
	if amount0.IsPositive() {
		// Use Amount0 to swap any Token1 ticks < (-depositTick0 -1 )
		depositTickToken0 := -tickIndex + feeInt64
		// subtract 1 from limit price because we can have double-sided liquidity at the deposit tick
		// we don't want to swap through this liquidity
		limitPrice0 := types.MustCalcPrice(-depositTickToken0 - 1)
		tradePairID := types.MustNewTradePairID(pairID.Token0, pairID.Token1)

		swapAmountIn0, swapAmountOut1, orderFilled, err := k.PerformSwapOnDepositSwap(ctx, tradePairID, amount0, limitPrice0)
		if err != nil {
			return math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), err
		}

		if swapAmountIn0.IsPositive() {
			if orderFilled {
				// due to decimal imprecision, we may have tiny remainder
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

		swapAmountIn1, swapAmountOut0, orderFilled, err := k.PerformSwapOnDepositSwap(ctx, tradePairID, amount1, limitPrice1)
		if err != nil {
			return math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), err
		}

		if swapAmountIn1.IsPositive() {
			if swappedToken0 {
				// This should be impossible, but leaving this as an extra precaution
				return math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), math_utils.ZeroPrecDec(), types.ErrDoubleSidedSwapOnDeposit
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
