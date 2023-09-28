package keeper_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	dualityapp "github.com/neutron-org/neutron/app"
	"github.com/neutron-org/neutron/testutil"
	"github.com/neutron-org/neutron/x/dex/types"
	"github.com/stretchr/testify/suite"
)

// Test Suite ///////////////////////////////////////////////////////////////
type CoreHelpersTestSuite struct {
	suite.Suite
	app   *dualityapp.App
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
	app := testutil.Setup(s.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	app.BankKeeper.SetParams(ctx, banktypes.DefaultParams())

	accAlice := app.AccountKeeper.NewAccountWithAddress(ctx, s.alice)
	app.AccountKeeper.SetAccount(ctx, accAlice)
	accBob := app.AccountKeeper.NewAccountWithAddress(ctx, s.bob)
	app.AccountKeeper.SetAccount(ctx, accBob)
	accCarol := app.AccountKeeper.NewAccountWithAddress(ctx, s.carol)
	app.AccountKeeper.SetAccount(ctx, accCarol)
	accDan := app.AccountKeeper.NewAccountWithAddress(ctx, s.dan)
	app.AccountKeeper.SetAccount(ctx, accDan)

	s.app = app
	s.ctx = ctx
	s.alice = sdk.AccAddress([]byte("alice"))
	s.bob = sdk.AccAddress([]byte("bob"))
	s.carol = sdk.AccAddress([]byte("carol"))
	s.dan = sdk.AccAddress([]byte("dan"))
}

func (s *CoreHelpersTestSuite) setLPAtFee1Pool(tickIndex int64, amountA int, amountB int) {
	pairID := &types.PairID{Token0: "TokenA", Token1: "TokenB"}
	pool, err := s.app.DexKeeper.GetOrInitPool(s.ctx, pairID, tickIndex, 1)
	if err != nil {
		panic(err)
	}
	lowerTick, upperTick := pool.LowerTick0, pool.UpperTick1
	amountAInt := sdk.NewInt(int64(amountA))
	amountBInt := sdk.NewInt(int64(amountB))

	existingShares := s.app.BankKeeper.GetSupply(s.ctx, pool.GetPoolDenom()).Amount

	totalShares := pool.CalcSharesMinted(amountAInt, amountBInt, existingShares)

	s.app.DexKeeper.MintShares(s.ctx, s.alice, totalShares)

	lowerTick.ReservesMakerDenom = amountAInt
	upperTick.ReservesMakerDenom = amountBInt
	s.app.DexKeeper.SetPool(s.ctx, pool)
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
	s.setLPAtFee1Pool(1, 0, 0)

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
	s.setLPAtFee1Pool(-1, 0, 0)
	s.setLPAtFee1Pool(1, 0, 10)

	// THEN GetCurrTick0To1 finds the tick at 2

	tickIdx, found := s.app.DexKeeper.GetCurrTickIndexTakerToMakerNormalized(s.ctx, defaultTradePairID0To1)
	s.Require().True(found)
	s.Assert().Equal(int64(2), tickIdx)
}
