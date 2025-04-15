package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	neutronapp "github.com/neutron-org/neutron/v6/app"
	"github.com/neutron-org/neutron/v6/testutil"
	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

// Test Suite ///////////////////////////////////////////////////////////////
type CoreHelpersTestSuite struct {
	suite.Suite
	app   *neutronapp.App
	ctx   sdk.Context
	alice sdk.AccAddress
	bob   sdk.AccAddress
	carol sdk.AccAddress
	dan   sdk.AccAddress
}

func TestCoreHelpersTestSuite(t *testing.T) {
	suite.Run(t, new(CoreHelpersTestSuite))
}

func (s *CoreHelpersTestSuite) SetupTest() {
	app := testutil.Setup(s.T())
	// `NewUncachedContext` like a `NewContext` calls `sdk.NewContext` under the hood. But the reason why we switched to NewUncachedContext
	// is NewContext tries to pass `app.finalizeBlockState.ms` as first argument while  app.finalizeBlockState is nil at this stage,
	// and we get nil pointer exception
	// when NewUncachedContext passes `app.cms` (multistore) as an argument to `sdk.NewContext`
	ctx := app.(*neutronapp.App).BaseApp.NewUncachedContext(false, cmtproto.Header{})

	accAlice := app.(*neutronapp.App).AccountKeeper.NewAccountWithAddress(ctx, s.alice)
	app.(*neutronapp.App).AccountKeeper.SetAccount(ctx, accAlice)
	accBob := app.(*neutronapp.App).AccountKeeper.NewAccountWithAddress(ctx, s.bob)
	app.(*neutronapp.App).AccountKeeper.SetAccount(ctx, accBob)
	accCarol := app.(*neutronapp.App).AccountKeeper.NewAccountWithAddress(ctx, s.carol)
	app.(*neutronapp.App).AccountKeeper.SetAccount(ctx, accCarol)
	accDan := app.(*neutronapp.App).AccountKeeper.NewAccountWithAddress(ctx, s.dan)
	app.(*neutronapp.App).AccountKeeper.SetAccount(ctx, accDan)

	s.app = app.(*neutronapp.App)
	s.ctx = ctx
	s.alice = []byte("alice")
	s.bob = []byte("bob")
	s.carol = []byte("carol")
	s.dan = []byte("dan")
}

func (s *CoreHelpersTestSuite) setLPAtFee1Pool(tickIndex int64, amountA, amountB int) {
	pairID := &types.PairID{Token0: "TokenA", Token1: "TokenB"}
	pool, err := s.app.DexKeeper.GetOrInitPool(s.ctx, pairID, tickIndex, 1)
	if err != nil {
		panic(err)
	}
	lowerTick, upperTick := pool.LowerTick0, pool.UpperTick1
	amountAInt := math.NewInt(int64(amountA))
	amountBInt := math.NewInt(int64(amountB))

	existingShares := s.app.BankKeeper.GetSupply(s.ctx, pool.GetPoolDenom()).Amount

	depositAmountAsToken0 := types.CalcAmountAsToken0(amountAInt, amountBInt, pool.MustCalcPrice1To0Center())
	totalShares := pool.CalcSharesMinted(depositAmountAsToken0, existingShares, math_utils.ZeroPrecDec())

	err = s.app.DexKeeper.MintShares(s.ctx, s.alice, sdk.NewCoins(totalShares))
	s.Require().NoError(err)

	lowerTick.ReservesMakerDenom = amountAInt
	upperTick.ReservesMakerDenom = amountBInt
	s.app.DexKeeper.UpdatePool(s.ctx, pool)
}

// FindNextTick ////////////////////////////////////////////////////

func (s *CoreHelpersTestSuite) TestFindNextTick1To0NoLiq() {
	// GIVEN there is no ticks with token0 in the pool

	s.setLPAtFee1Pool(1, 0, 10)

	// THEN GetCurrTick1To0 doesn't find a tick

	_, found := s.app.DexKeeper.GetCurrTickIndexTakerToMaker(s.ctx, defaultTradePairID1To0)
	s.Assert().False(found)
}

func (s *CoreHelpersTestSuite) TestGetCurrTick1To0WithLiq() {
	// Given multiple locations of token0
	s.setLPAtFee1Pool(-1, 10, 0)
	s.setLPAtFee1Pool(0, 10, 0)

	// THEN GetCurrTick1To0 finds the tick at -1

	tickIdx, found := s.app.DexKeeper.GetCurrTickIndexTakerToMakerNormalized(s.ctx, defaultTradePairID1To0)
	s.Require().True(found)
	s.Assert().Equal(int64(-1), tickIdx)
}

func (s *CoreHelpersTestSuite) TestGetCurrTick1To0WithMinLiq() {
	// GIVEN tick with token0 @ index -1
	s.setLPAtFee1Pool(-1, 10, 0)

	// THEN GetCurrTick1To0 finds the tick at -2

	tickIdx, found := s.app.DexKeeper.GetCurrTickIndexTakerToMakerNormalized(s.ctx, defaultTradePairID1To0)
	s.Require().True(found)
	s.Assert().Equal(int64(-2), tickIdx)
}

// GetCurrTick0To1 ///////////////////////////////////////////////////////////

func (s *CoreHelpersTestSuite) TestGetCurrTick0To1NoLiq() {
	// GIVEN there are no tick with Token1 in the pool

	s.setLPAtFee1Pool(0, 10, 0)

	// THEN GetCurrTick0To1 doesn't find a tick

	_, found := s.app.DexKeeper.GetCurrTickIndexTakerToMakerNormalized(s.ctx, defaultTradePairID0To1)
	s.Assert().False(found)
}

func (s *CoreHelpersTestSuite) TestGetCurrTick0To1WithLiq() {
	// GIVEN multiple locations of token1

	s.setLPAtFee1Pool(-1, 10, 0)
	s.setLPAtFee1Pool(0, 0, 10)
	s.setLPAtFee1Pool(1, 0, 10)

	// THEN GetCurrTick0To1 finds the tick at 1

	tickIdx, found := s.app.DexKeeper.GetCurrTickIndexTakerToMakerNormalized(s.ctx, defaultTradePairID0To1)
	s.Require().True(found)
	s.Assert().Equal(int64(1), tickIdx)
}

func (s *CoreHelpersTestSuite) TestGetCurrTick0To1WithMinLiq() {
	// WHEN tick with token1 @ index 1
	s.setLPAtFee1Pool(1, 0, 10)

	// THEN GetCurrTick0To1 finds the tick at 2

	tickIdx, found := s.app.DexKeeper.GetCurrTickIndexTakerToMakerNormalized(s.ctx, defaultTradePairID0To1)
	s.Require().True(found)
	s.Assert().Equal(int64(2), tickIdx)
}

// IsBehindEnemyLines /////////////////////////////////////////////////////////

func (s *CoreHelpersTestSuite) TestIsBehindEnemyLinesToken0BELHighTick() {
	s.setLPAtFee1Pool(100, 0, 10)
	tradePairID := types.MustNewTradePairID("TokenB", "TokenA")
	isBEL := s.app.DexKeeper.IsBehindEnemyLines(s.ctx, tradePairID, -102)
	s.True(isBEL)
}

func (s *CoreHelpersTestSuite) TestIsBehindEnemyLinesToken0BELLowTick() {
	s.setLPAtFee1Pool(-100, 0, 10)
	tradePairID := types.MustNewTradePairID("TokenB", "TokenA")
	isBEL := s.app.DexKeeper.IsBehindEnemyLines(s.ctx, tradePairID, 98)
	s.True(isBEL)
}

func (s *CoreHelpersTestSuite) TestIsBehindEnemyLinesToken0ValidHighTick() {
	s.setLPAtFee1Pool(100, 0, 10)
	tradePairID := types.MustNewTradePairID("TokenB", "TokenA")
	isBEL := s.app.DexKeeper.IsBehindEnemyLines(s.ctx, tradePairID, -101)
	s.False(isBEL)
}

func (s *CoreHelpersTestSuite) TestIsBehindEnemyLinesToken0ValidLowtick() {
	s.setLPAtFee1Pool(-100, 0, 10)
	tradePairID := types.MustNewTradePairID("TokenB", "TokenA")
	isBEL := s.app.DexKeeper.IsBehindEnemyLines(s.ctx, tradePairID, 99)
	s.False(isBEL)
}

func (s *CoreHelpersTestSuite) TestIsBehindEnemyLinesToken1BELLowTick() {
	s.setLPAtFee1Pool(-10, 10, 0)
	tradePairID := types.MustNewTradePairID("TokenA", "TokenB")
	isBEL := s.app.DexKeeper.IsBehindEnemyLines(s.ctx, tradePairID, -12)
	s.True(isBEL)
}

func (s *CoreHelpersTestSuite) TestIsBehindEnemyLinesToken1BELHighTick() {
	s.setLPAtFee1Pool(10, 10, 0)
	tradePairID := types.MustNewTradePairID("TokenA", "TokenB")
	isBEL := s.app.DexKeeper.IsBehindEnemyLines(s.ctx, tradePairID, 8)
	s.True(isBEL)
}

func (s *CoreHelpersTestSuite) TestIsBehindEnemyLinesToken1ValidLowTick() {
	s.setLPAtFee1Pool(-10, 10, 0)
	tradePairID := types.MustNewTradePairID("TokenA", "TokenB")
	isBEL := s.app.DexKeeper.IsBehindEnemyLines(s.ctx, tradePairID, -11)
	s.False(isBEL)
}

func (s *CoreHelpersTestSuite) TestIsBehindEnemyLinesToken1ValidHighTick() {
	s.setLPAtFee1Pool(10, 10, 0)
	tradePairID := types.MustNewTradePairID("TokenA", "TokenB")
	isBEL := s.app.DexKeeper.IsBehindEnemyLines(s.ctx, tradePairID, 9)
	s.False(isBEL)
}

func (s *CoreHelpersTestSuite) TestIsPoolBehindEnemyLinesToken0BEL() {
	// GIVEN Token1 at tick -19
	s.setLPAtFee1Pool(-20, 0, 10)

	// WHEN create pool with token0 at 18
	isBEL := s.app.DexKeeper.IsPoolBehindEnemyLines(s.ctx, defaultPairID, -17, 1, math.OneInt(), math.ZeroInt())

	// THEN pool is BEL
	s.True(isBEL)
}

func (s *CoreHelpersTestSuite) TestIsPoolBehindEnemyLinesToken0Valid() {
	// GIVEN Token1 at tick -19
	s.setLPAtFee1Pool(-20, 0, 10)

	// WHEN create pool with token0 at 19
	isBEL := s.app.DexKeeper.IsPoolBehindEnemyLines(s.ctx, defaultPairID, -18, 1, math.OneInt(), math.ZeroInt())
	// THEN pool is not BEL
	s.False(isBEL)
}

func (s *CoreHelpersTestSuite) TestIsPoolBehindEnemyLinesToken1BEL() {
	// GIVEN token0 at -19
	s.setLPAtFee1Pool(20, 10, 0)

	// WHEN create pool with Token1 at 18
	isBEL := s.app.DexKeeper.IsPoolBehindEnemyLines(s.ctx, defaultPairID, 17, 1, math.ZeroInt(), math.OneInt())

	// THEN pool is BEL
	s.True(isBEL)
}

func (s *CoreHelpersTestSuite) TestIsPoolBehindEnemyLinesToken1Valid() {
	// GIVEN token0 at -19
	s.setLPAtFee1Pool(20, 10, 0)

	// WHEN create pool with Token1 at 19
	isBEL := s.app.DexKeeper.IsPoolBehindEnemyLines(s.ctx, defaultPairID, 18, 1, math.ZeroInt(), math.OneInt())

	// THEN pool is BEL
	s.False(isBEL)
}

func (s *CoreHelpersTestSuite) TestExpiredLimitOrderNotCountedForBEL() {
	s.ctx = s.ctx.WithBlockTime(time.Now())
	// Given a GTT tranche that has not expired
	expTime := s.ctx.BlockTime().Add(time.Hour * 24)
	tranche := &types.LimitOrderTranche{
		Key: &types.LimitOrderTrancheKey{
			TradePairId:           defaultTradePairID1To0,
			TrancheKey:            "1",
			TickIndexTakerToMaker: 0,
		},
		ExpirationTime:     &expTime,
		ReservesMakerDenom: math.NewInt(1000000),
	}

	s.app.DexKeeper.SetLimitOrderTranche(s.ctx, tranche)
	// Pool is behind enemy lines
	isBEL := s.app.DexKeeper.IsPoolBehindEnemyLines(s.ctx, defaultPairID, -5, 1, math.ZeroInt(), math.OneInt())
	s.True(isBEL)

	// When tranche is  expired
	expTime = s.ctx.BlockTime().Add(-time.Hour * 24)
	tranche.ExpirationTime = &expTime
	s.app.DexKeeper.UpdateTranche(s.ctx, tranche)
	isBEL = s.app.DexKeeper.IsPoolBehindEnemyLines(s.ctx, defaultPairID, -5, 1, math.ZeroInt(), math.OneInt())
	s.False(isBEL)
}
