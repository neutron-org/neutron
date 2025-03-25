package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

type LiquidityIterator struct {
	keeper      *Keeper
	tradePairID *types.TradePairID
	ctx         sdk.Context
	iter        TickIterator
}

func (k Keeper) NewLiquidityIterator(
	ctx sdk.Context,
	tradePairID *types.TradePairID,
) *LiquidityIterator {
	return &LiquidityIterator{
		iter:        k.NewTickIterator(ctx, tradePairID),
		keeper:      &k,
		ctx:         ctx,
		tradePairID: tradePairID,
	}
}

func (s *LiquidityIterator) Next() types.Liquidity {
	// Move iterator to the next tick after each call
	// iter must be in valid state to call next
	defer func() {
		if s.iter.Valid() {
			s.iter.Next()
		}
	}()

	for ; s.iter.Valid(); s.iter.Next() {
		tick := s.iter.Value()
		// Don't bother to look up pool counter-liquidities if there's no liquidity here
		if !tick.HasToken() {
			continue
		}

		liq := s.WrapTickLiquidity(tick)
		if liq != nil {
			return liq
		}
	}

	return nil
}

func (s *LiquidityIterator) Close() {
	s.iter.Close()
}

func (s *LiquidityIterator) WrapTickLiquidity(tick types.TickLiquidity) types.Liquidity {
	switch liquidity := tick.Liquidity.(type) {
	case *types.TickLiquidity_PoolReserves:
		poolReserves := liquidity.PoolReserves
		counterpartKey := poolReserves.Key.Counterpart()
		counterpartReserves, counterpartReservesFound := s.keeper.GetPoolReserves(s.ctx, counterpartKey)
		if !counterpartReservesFound {
			counterpartReserves = types.NewPoolReservesFromCounterpart(poolReserves)
		}

		var lowerTick0, upperTick1 *types.PoolReserves
		if s.tradePairID.IsTakerDenomToken0() {
			lowerTick0 = counterpartReserves
			upperTick1 = poolReserves
		} else {
			lowerTick0 = poolReserves
			upperTick1 = counterpartReserves
		}
		return &types.PoolLiquidity{
			TradePairID: s.tradePairID,
			Pool: &types.Pool{
				LowerTick0: lowerTick0,
				UpperTick1: upperTick1,
			},
		}

	case *types.TickLiquidity_LimitOrderTranche:
		tranche := liquidity.LimitOrderTranche
		// If we hit a tranche with an expired goodTil date keep iterating
		if tranche.IsExpired(s.ctx) {
			return nil
		}

		return tranche

	default:
		panic("Tick does not have liquidity")
	}
}
