package v601_test

import (
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	time "time"

	v601 "github.com/neutron-org/neutron/v6/app/upgrades/v6.0.1"
	"github.com/neutron-org/neutron/v6/testutil"
	dextypes "github.com/neutron-org/neutron/v6/x/dex/types"
)

type UpgradeTestSuite struct {
	testutil.IBCConnectionTestSuite
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (suite *UpgradeTestSuite) SetupTest() {
	suite.IBCConnectionTestSuite.SetupTest()
}

func (suite *UpgradeTestSuite) TestUpgradeRemoveOrphanedLimitOrders() {
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext().WithChainID("neutron-1")
	t := suite.T()

	upgrade := upgradetypes.Plan{
		Name:   v601.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}

	pairID := dextypes.MustNewTradePairID("TokenA", "TokenB")
	expTime := time.Now().Add(time.Hour * 24)

	// Create a couple orphaned GTT Tranches
	app.DexKeeper.SetLimitOrderTranche(ctx, &dextypes.LimitOrderTranche{
		Key: &dextypes.LimitOrderTrancheKey{
			TradePairId:           pairID,
			TrancheKey:            "1",
			TickIndexTakerToMaker: 0,
		},
		ExpirationTime: &expTime,
	},
	)
	app.DexKeeper.SetLimitOrderTranche(ctx, &dextypes.LimitOrderTranche{
		Key: &dextypes.LimitOrderTrancheKey{
			TradePairId:           pairID,
			TrancheKey:            "2",
			TickIndexTakerToMaker: 0,
		},
		ExpirationTime: &expTime,
	},
	)

	//Create an GTT tranche with an ExpirationRecord
	tranche := &dextypes.LimitOrderTranche{
		Key: &dextypes.LimitOrderTrancheKey{
			TradePairId:           pairID,
			TrancheKey:            "3",
			TickIndexTakerToMaker: 0,
		},
		ExpirationTime: &expTime,
	}
	app.DexKeeper.SetLimitOrderTranche(ctx, tranche)
	app.DexKeeper.SetLimitOrderExpiration(ctx, &dextypes.LimitOrderExpiration{
		TrancheRef:     tranche.Key.KeyMarshal(),
		ExpirationTime: expTime,
	})

	// Create a Normal Tranche
	tranche = &dextypes.LimitOrderTranche{
		Key: &dextypes.LimitOrderTrancheKey{
			TradePairId:           pairID,
			TrancheKey:            "4",
			TickIndexTakerToMaker: 0,
		},
	}
	app.DexKeeper.SetLimitOrderTranche(ctx, tranche)
	require.NoError(t, app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))

	tickLiquidity := app.DexKeeper.GetAllTickLiquidity(ctx)

	// Only the healthy tranches remain
	require.Equal(t, len(tickLiquidity), 2)
	require.Equal(t, tickLiquidity[0].GetLimitOrderTranche().Key.TrancheKey, "3")
	require.Equal(t, tickLiquidity[1].GetLimitOrderTranche().Key.TrancheKey, "4")

	// The expired tranches are converted to inactiveLimitOrderTranches
	inactiveTranche := app.DexKeeper.GetAllInactiveLimitOrderTranche(ctx)
	require.Equal(t, len(inactiveTranche), 2)
	require.Equal(t, inactiveTranche[0].Key.TrancheKey, "1")
	require.Equal(t, inactiveTranche[1].Key.TrancheKey, "2")
}
