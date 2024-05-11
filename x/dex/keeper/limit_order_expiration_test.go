package keeper_test

import (
	"strconv"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v3/x/dex/keeper"
	"github.com/neutron-org/neutron/v3/x/dex/types"
)

func createNLimitOrderExpiration(
	keeper *keeper.Keeper,
	ctx sdk.Context,
	n int,
) []types.LimitOrderExpiration {
	items := make([]types.LimitOrderExpiration, n)
	for i := range items {
		items[i].ExpirationTime = time.Unix(int64(i), 10).UTC()
		items[i].TrancheRef = []byte(strconv.Itoa(i))

		keeper.SetLimitOrderExpiration(ctx, &items[i])
	}

	return items
}

func createLimitOrderExpirationAndTranches(
	keeper *keeper.Keeper,
	ctx sdk.Context,
	expTimes []time.Time,
) {
	items := make([]types.LimitOrderExpiration, len(expTimes))
	for i := range items {
		tranche := &types.LimitOrderTranche{
			Key: &types.LimitOrderTrancheKey{
				TradePairId: &types.TradePairID{
					MakerDenom: "TokenA",
					TakerDenom: "TokenB",
				},
				TickIndexTakerToMaker: 0,
				TrancheKey:            strconv.Itoa(i),
			},
			ReservesMakerDenom: math.NewInt(10),
			ReservesTakerDenom: math.NewInt(10),
			TotalMakerDenom:    math.NewInt(10),
			TotalTakerDenom:    math.NewInt(10),
			ExpirationTime:     &expTimes[i],
		}
		items[i].ExpirationTime = expTimes[i]
		items[i].TrancheRef = tranche.Key.KeyMarshal()

		keeper.SetLimitOrderExpiration(ctx, &items[i])
		keeper.SetLimitOrderTranche(ctx, tranche)
	}
}

func (s *DexTestSuite) TestLimitOrderExpirationGet() {
	keeper := s.App.DexKeeper
	items := createNLimitOrderExpiration(&keeper, s.Ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetLimitOrderExpiration(s.Ctx,
			item.ExpirationTime,
			item.TrancheRef,
		)
		s.True(found)
		s.Equal(item, *rst)
	}
}

func (s *DexTestSuite) TestLimitOrderExpirationRemove() {
	keeper := s.App.DexKeeper
	items := createNLimitOrderExpiration(&keeper, s.Ctx, 10)
	for _, item := range items {
		keeper.RemoveLimitOrderExpiration(s.Ctx,
			item.ExpirationTime,
			item.TrancheRef,
		)
		_, found := keeper.GetLimitOrderExpiration(s.Ctx,
			item.ExpirationTime,
			item.TrancheRef,
		)
		s.False(found)
	}
}

func (s *DexTestSuite) TestLimitOrderExpirationGetAll() {
	items := createNLimitOrderExpiration(&s.App.DexKeeper, s.Ctx, 10)
	pointerItems := make([]*types.LimitOrderExpiration, len(items))
	for i := range items {
		pointerItems[i] = &items[i]
	}
	s.ElementsMatch(
		pointerItems,
		s.App.DexKeeper.GetAllLimitOrderExpiration(s.Ctx),
	)
}

func (s *DexTestSuite) TestPurgeExpiredLimitOrders() {
	keeper := s.App.DexKeeper
	now := time.Now().UTC()
	ctx := s.Ctx.WithBlockTime(now)
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(1000000))

	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)
	nextWeek := now.AddDate(0, 0, 7)

	expTimes := []time.Time{
		yesterday,
		yesterday,
		now,
		tomorrow,
		nextWeek,
	}

	createLimitOrderExpirationAndTranches(&keeper, s.Ctx, expTimes)
	keeper.PurgeExpiredLimitOrders(s.Ctx, now)

	// Only future LimitOrderExpiration items still exist
	expList := keeper.GetAllLimitOrderExpiration(s.Ctx)
	s.Equal(2, len(expList))
	s.Equal(tomorrow, expList[0].ExpirationTime)
	s.Equal(nextWeek, expList[1].ExpirationTime)

	// Only future LimitOrderTranches Exist
	trancheList := keeper.GetAllLimitOrderTrancheAtIndex(s.Ctx, defaultTradePairID1To0, 0)
	s.Equal(2, len(trancheList))
	s.Equal(tomorrow, *trancheList[0].ExpirationTime)
	s.Equal(nextWeek, *trancheList[1].ExpirationTime)

	// InactiveLimitOrderTranches have been created for the expired tranched
	inactiveTrancheList := keeper.GetAllInactiveLimitOrderTranche(ctx)
	s.Equal(3, len(inactiveTrancheList))
	s.Equal(yesterday, *inactiveTrancheList[0].ExpirationTime)
	s.Equal(yesterday, *inactiveTrancheList[1].ExpirationTime)
	s.Equal(now, *inactiveTrancheList[2].ExpirationTime)
}

func (s *DexTestSuite) TestPurgeExpiredLimitOrdersAtBlockGasLimit() {
	keeper := s.App.DexKeeper
	now := time.Now().UTC()
	ctx := s.Ctx.WithBlockTime(now)
	gasLimit := 1000000
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(uint64(gasLimit)))
	timeRequiredToPurgeOneNonJIT := 35000
	gasUsed := gasLimit - types.GoodTilPurgeGasBuffer - timeRequiredToPurgeOneNonJIT

	yesterday := now.AddDate(0, 0, -1)

	expTimes := []time.Time{
		types.JITGoodTilTime(),
		types.JITGoodTilTime(),
		yesterday,
		yesterday,
		yesterday,
	}
	createLimitOrderExpirationAndTranches(&keeper, ctx, expTimes)

	// IF blockGasMeter is nearing the GoodTilPurgeBuffer
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(uint64(gasLimit)))
	ctx.BlockGasMeter().ConsumeGas(uint64(gasUsed), "stub block gas usage")

	// WHEN PurgeExpiredLimitOrders is run
	keeper.PurgeExpiredLimitOrders(ctx, now)

	// THEN GoodTilPurgeHitGasLimit event is emitted
	s.AssertEventEmitted(ctx, types.EventTypeGoodTilPurgeHitGasLimit, 1)

	// All JIT expirations are purged but other expirations remain
	expList := keeper.GetAllLimitOrderExpiration(ctx)
	// NOTE: this test is very brittle because it relies on an estimated cost
	// for deleting expirations. If this cost changes the number of remaining
	// expirations may change
	s.Equal(1, len(expList))
	s.Equal(expList[0].ExpirationTime, yesterday)
}

func (s *DexTestSuite) TestPurgeExpiredLimitOrdersAtBlockGasLimitOnlyJIT() {
	keeper := s.App.DexKeeper
	now := time.Now().UTC()
	ctx := s.Ctx.WithBlockTime(now)
	gasLimt := 1000000
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(uint64(gasLimt)))
	gasUsed := gasLimt - types.GoodTilPurgeGasBuffer - 30000

	expTimes := []time.Time{
		types.JITGoodTilTime(),
		types.JITGoodTilTime(),
		types.JITGoodTilTime(),
		types.JITGoodTilTime(),
		types.JITGoodTilTime(),
		types.JITGoodTilTime(),
		types.JITGoodTilTime(),
	}

	createLimitOrderExpirationAndTranches(&keeper, ctx, expTimes)
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(100000))
	ctx.BlockGasMeter().ConsumeGas(uint64(gasUsed), "stub block gas usage")
	keeper.PurgeExpiredLimitOrders(ctx, now)

	// GoodTilPurgeHitGasLimit event is not been emitted
	s.AssertEventValueNotEmitted(types.GoodTilPurgeHitGasLimitEventGas, "Hit gas limit purging JIT expirations")

	// All JIT expirations are purged
	expList := keeper.GetAllLimitOrderExpiration(ctx)
	s.Equal(0, len(expList))
}
