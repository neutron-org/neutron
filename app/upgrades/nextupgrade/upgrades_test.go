package nextupgrade_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/neutron-org/neutron/v11/app/upgrades/nextupgrade"
	"github.com/neutron-org/neutron/v11/testutil"
	dextypes "github.com/neutron-org/neutron/v11/x/dex/types"
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

func (suite *UpgradeTestSuite) TestUpgrade() {
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext().WithChainID("neutron-1")
	t := suite.T()

	upgrade := upgradetypes.Plan{
		Name:   nextupgrade.UpgradeName,
		Info:   "some text here",
		Height: 100,
	}

	require.NoError(t, app.UpgradeKeeper.ApplyUpgrade(ctx, upgrade))
}

// TestReconstructLoExpirations verifies the LimitOrderExpiration migration.
//
// A LimitOrderExpiration stores a TrancheRef = tranche.Key.KeyMarshal(), which embeds the
// TrancheKey string as part of its bytes. When a tranche key is rewritten from the old
// plain-decimal "tk-N" format to the zero-padded "tk-%020d" format, the corresponding
// expiration entry must be removed from its old store key and re-inserted under the new one.
//
// Entries pointing to a base-36 tranche key (no "tk-" prefix) must be left untouched.
func (suite *UpgradeTestSuite) TestReconstructLoExpirations() {
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext().WithChainID("neutron-1")
	t := suite.T()

	expTime1 := time.Now().UTC().Add(time.Second * 111)
	expTime2 := time.Now().UTC().Add(time.Second * 222)
	expTime3 := time.Now().UTC().Add(time.Second * 333)

	// ── pre-upgrade state ────────────────────────────────────────────────────

	// Two active limit-order tranches with old plain-decimal keys and different ExpirationTime.
	pairID1 := dextypes.MustNewTradePairID(
		"ibc/B559A80D62249C8AA07A380E2A2BEA6E5CA9A6F079C912C3A9E9B494105E4F81",
		"factory/neutron1frc0p5czd9uaaymdkug2njz7dc7j65jxukp9apmt9260a8egujkspms2t2/udntrn",
	)
	oldKey1 := &dextypes.LimitOrderTrancheKey{
		TradePairId:           pairID1,
		TrancheKey:            "tk-19993998",
		TickIndexTakerToMaker: -43028,
	}
	app.DexKeeper.SetLimitOrderTranche(ctx, &dextypes.LimitOrderTranche{
		Key:            oldKey1,
		ExpirationTime: &expTime1,
	})
	app.DexKeeper.SetLimitOrderExpiration(ctx, &dextypes.LimitOrderExpiration{
		ExpirationTime: expTime1,
		TrancheRef:     oldKey1.KeyMarshal(),
	})

	oldKey2 := &dextypes.LimitOrderTrancheKey{
		TradePairId:           pairID1,
		TrancheKey:            "tk-19940606",
		TickIndexTakerToMaker: -42321,
	}
	app.DexKeeper.SetLimitOrderTranche(ctx, &dextypes.LimitOrderTranche{
		Key:            oldKey2,
		ExpirationTime: &expTime2,
	})
	app.DexKeeper.SetLimitOrderExpiration(ctx, &dextypes.LimitOrderExpiration{
		ExpirationTime: expTime2,
		TrancheRef:     oldKey2.KeyMarshal(),
	})

	// One tranche with the original base-36 sortable key (no "tk-" prefix).
	// Its expiration must not be touched.
	pairID2 := dextypes.MustNewTradePairID(
		"factory/neutron1dqd0wsqldr89m4d9trk2arv35twz7a5erjj6td/nick",
		"factory/neutron1dqd0wsqldr89m4d9trk2arv35twz7a5erjj6td/jcp",
	)
	base36Key := &dextypes.LimitOrderTrancheKey{
		TradePairId:           pairID2,
		TrancheKey:            "57mgzl47if5",
		TickIndexTakerToMaker: 46055,
	}
	app.DexKeeper.SetLimitOrderTranche(ctx, &dextypes.LimitOrderTranche{
		Key:            base36Key,
		ExpirationTime: &expTime3,
	})
	app.DexKeeper.SetLimitOrderExpiration(ctx, &dextypes.LimitOrderExpiration{
		ExpirationTime: expTime3,
		TrancheRef:     base36Key.KeyMarshal(),
	})

	require.Len(t, app.DexKeeper.GetAllLimitOrderExpiration(ctx), 3, "pre-upgrade: 3 expirations")

	// ── run migration ────────────────────────────────────────────────────────

	require.NoError(t, nextupgrade.ReconstructTrancheKeys(ctx, app.AppCodec(), app.DexKeeper))

	// ── post-upgrade assertions ──────────────────────────────────────────────

	require.Len(t, app.DexKeeper.GetAllLimitOrderExpiration(ctx), 3, "post-upgrade: expiration count must not change")

	// --- tk-19993998 expiration → new TrancheRef under tk-00000000000019993998 ---

	newKey1 := &dextypes.LimitOrderTrancheKey{
		TradePairId:           pairID1,
		TrancheKey:            dextypes.NewTrancheKey(19993998),
		TickIndexTakerToMaker: -43028,
	}
	_, found := app.DexKeeper.GetLimitOrderExpiration(ctx, expTime1, newKey1.KeyMarshal())
	require.True(t, found, "new expiration (tk-00000000000019993998) must exist")
	_, found = app.DexKeeper.GetLimitOrderExpiration(ctx, expTime1, oldKey1.KeyMarshal())
	require.False(t, found, "old expiration (tk-19993998) must be removed")

	// --- tk-19940606 expiration → new TrancheRef under tk-00000000000019940606 ---

	newKey2 := &dextypes.LimitOrderTrancheKey{
		TradePairId:           pairID1,
		TrancheKey:            dextypes.NewTrancheKey(19940606),
		TickIndexTakerToMaker: -42321,
	}
	_, found = app.DexKeeper.GetLimitOrderExpiration(ctx, expTime2, newKey2.KeyMarshal())
	require.True(t, found, "new expiration (tk-00000000000019940606) must exist")
	_, found = app.DexKeeper.GetLimitOrderExpiration(ctx, expTime2, oldKey2.KeyMarshal())
	require.False(t, found, "old expiration (tk-19940606) must be removed")

	// --- base-36 expiration must be unchanged ---

	_, found = app.DexKeeper.GetLimitOrderExpiration(ctx, expTime3, base36Key.KeyMarshal())
	require.True(t, found, "base-36 expiration must still exist")
}

// TestReconstructLoTrancheKeys verifies the tranche key migration from the plain-decimal
// "tk-N" format to the zero-padded "tk-%020d" format. It uses realistic mainnet entries:
//
//   - tk-19993998 / tk-19940606: old decimal keys that must be rewritten.
//   - 57mgzl47if5: original base-36 sortable key (no "tk-" prefix) that must be left untouched.
//   - pool_reserves: a tick-liquidity entry that is not a limit order; must be untouched.
func (suite *UpgradeTestSuite) TestReconstructLoTrancheKeys() {
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext().WithChainID("neutron-1")
	t := suite.T()

	// ── pre-upgrade state ────────────────────────────────────────────────────

	// Two active limit-order tranches with old plain-decimal keys.
	pairID1 := dextypes.MustNewTradePairID(
		"ibc/B559A80D62249C8AA07A380E2A2BEA6E5CA9A6F079C912C3A9E9B494105E4F81",
		"factory/neutron1frc0p5czd9uaaymdkug2njz7dc7j65jxukp9apmt9260a8egujkspms2t2/udntrn",
	)
	app.DexKeeper.SetLimitOrderTranche(ctx, &dextypes.LimitOrderTranche{
		Key: &dextypes.LimitOrderTrancheKey{
			TradePairId:           pairID1,
			TrancheKey:            "tk-19993998",
			TickIndexTakerToMaker: -43028,
		},
	})
	app.DexKeeper.SetLimitOrderTranche(ctx, &dextypes.LimitOrderTranche{
		Key: &dextypes.LimitOrderTrancheKey{
			TradePairId:           pairID1,
			TrancheKey:            "tk-19940606",
			TickIndexTakerToMaker: -42321,
		},
	})

	// One tranche with the original base-36 sortable key (no "tk-" prefix).
	// The migration must skip it and leave it unchanged.
	pairID2 := dextypes.MustNewTradePairID(
		"factory/neutron1dqd0wsqldr89m4d9trk2arv35twz7a5erjj6td/nick",
		"factory/neutron1dqd0wsqldr89m4d9trk2arv35twz7a5erjj6td/jcp",
	)
	app.DexKeeper.SetLimitOrderTranche(ctx, &dextypes.LimitOrderTranche{
		Key: &dextypes.LimitOrderTrancheKey{
			TradePairId:           pairID2,
			TrancheKey:            "57mgzl47if5",
			TickIndexTakerToMaker: 46055,
		},
	})

	// One pool-reserves entry. It is stored under the same TickLiquidity prefix but is
	// not a limit order; the migration must not touch it.
	pairID3 := dextypes.MustNewTradePairID(
		"ibc/E2A000FD3EDD91C9429B473995CE2C7C555BCC8CFC1D0A3D02F514392B7A80E8",
		"factory/neutron17sp75wng9vl2hu3sf4ky86d7smmk3wle9gkts2gmedn9x4ut3xcqa5xp34/maxbtc",
	)
	app.DexKeeper.SetPoolReserves(ctx, &dextypes.PoolReserves{
		Key: &dextypes.PoolReservesKey{
			TradePairId:           pairID3,
			TickIndexTakerToMaker: 187,
			Fee:                   102,
		},
	})

	require.Len(t, app.DexKeeper.GetAllTickLiquidity(ctx), 4, "pre-upgrade: 4 tick liquidity entries")

	// ── run migration ────────────────────────────────────────────────────────

	require.NoError(t, nextupgrade.ReconstructTrancheKeys(ctx, app.AppCodec(), app.DexKeeper))

	// ── post-upgrade assertions ──────────────────────────────────────────────

	// Total entry count must be unchanged.
	require.Len(t, app.DexKeeper.GetAllTickLiquidity(ctx), 4, "post-upgrade: entry count must not change")

	// --- tk-19993998 → tk-00000000000019993998 ---

	migratedKey1 := dextypes.NewTrancheKey(19993998) // "tk-00000000000019993998"

	migratedTranche1 := app.DexKeeper.GetLimitOrderTranche(ctx, &dextypes.LimitOrderTrancheKey{
		TradePairId:           pairID1,
		TrancheKey:            migratedKey1,
		TickIndexTakerToMaker: -43028,
	})
	require.NotNil(t, migratedTranche1, "migrated tranche tk-19993998 must exist under new key")
	require.Equal(t, migratedKey1, migratedTranche1.Key.TrancheKey)

	require.Nil(t,
		app.DexKeeper.GetLimitOrderTranche(ctx, &dextypes.LimitOrderTrancheKey{
			TradePairId:           pairID1,
			TrancheKey:            "tk-19993998",
			TickIndexTakerToMaker: -43028,
		}),
		"old key tk-19993998 must no longer exist",
	)

	// --- tk-19940606 → tk-00000000000019940606 ---

	migratedKey2 := dextypes.NewTrancheKey(19940606) // "tk-00000000000019940606"

	migratedTranche2 := app.DexKeeper.GetLimitOrderTranche(ctx, &dextypes.LimitOrderTrancheKey{
		TradePairId:           pairID1,
		TrancheKey:            migratedKey2,
		TickIndexTakerToMaker: -42321,
	})
	require.NotNil(t, migratedTranche2, "migrated tranche tk-19940606 must exist under new key")
	require.Equal(t, migratedKey2, migratedTranche2.Key.TrancheKey)

	require.Nil(t,
		app.DexKeeper.GetLimitOrderTranche(ctx, &dextypes.LimitOrderTrancheKey{
			TradePairId:           pairID1,
			TrancheKey:            "tk-19940606",
			TickIndexTakerToMaker: -42321,
		}),
		"old key tk-19940606 must no longer exist",
	)

	// --- 57mgzl47if5 (base-36 key, no "tk-" prefix) must be unchanged ---

	untouchedTranche := app.DexKeeper.GetLimitOrderTranche(ctx, &dextypes.LimitOrderTrancheKey{
		TradePairId:           pairID2,
		TrancheKey:            "57mgzl47if5",
		TickIndexTakerToMaker: 46055,
	})
	require.NotNil(t, untouchedTranche, "base-36 tranche must still exist")
	require.Equal(t, "57mgzl47if5", untouchedTranche.Key.TrancheKey, "base-36 key must not be rewritten")

	// --- pool_reserves must be untouched ---

	poolReserves, found := app.DexKeeper.GetPoolReserves(ctx, &dextypes.PoolReservesKey{
		TradePairId:           pairID3,
		TickIndexTakerToMaker: 187,
		Fee:                   102,
	})
	require.True(t, found, "pool reserves must still be present")
	require.NotNil(t, poolReserves)
}

// TestReconstructInactiveLoTranches verifies that inactive limit order tranches stored
// under the old plain-decimal "tk-N" key format are rewritten to the zero-padded
// "tk-%020d" format, while entries with the original base-36 sortable key are left alone.
func (suite *UpgradeTestSuite) TestReconstructInactiveLoTranches() {
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext().WithChainID("neutron-1")
	t := suite.T()

	// ── pre-upgrade state ────────────────────────────────────────────────────

	// One inactive tranche with the original base-36 sortable key (no "tk-" prefix).
	pairID1 := dextypes.MustNewTradePairID(
		"factory/neutron10h9stc5v6ntgeygf5xf945njqq5h32r54rf7kf/nick",
		"factory/neutron1dqd0wsqldr89m4d9trk2arv35twz7a5erjj6td/jcp",
	)
	app.DexKeeper.SetInactiveLimitOrderTranche(ctx, &dextypes.LimitOrderTranche{
		Key: &dextypes.LimitOrderTrancheKey{
			TradePairId:           pairID1,
			TrancheKey:            "57m0a14awvr",
			TickIndexTakerToMaker: 0,
		},
	})

	// Two inactive tranches with old plain-decimal keys.
	pairID2 := dextypes.MustNewTradePairID(
		"ibc/B559A80D62249C8AA07A380E2A2BEA6E5CA9A6F079C912C3A9E9B494105E4F81",
		"factory/neutron17sp75wng9vl2hu3sf4ky86d7smmk3wle9gkts2gmedn9x4ut3xcqa5xp34/maxbtc",
	)
	app.DexKeeper.SetInactiveLimitOrderTranche(ctx, &dextypes.LimitOrderTranche{
		Key: &dextypes.LimitOrderTrancheKey{
			TradePairId:           pairID2,
			TrancheKey:            "tk-18498162",
			TickIndexTakerToMaker: 67831,
		},
	})
	app.DexKeeper.SetInactiveLimitOrderTranche(ctx, &dextypes.LimitOrderTranche{
		Key: &dextypes.LimitOrderTrancheKey{
			TradePairId:           pairID2,
			TrancheKey:            "tk-18291522",
			TickIndexTakerToMaker: 67961,
		},
	})

	require.Len(t, app.DexKeeper.GetAllInactiveLimitOrderTranche(ctx), 3, "pre-upgrade: 3 inactive tranches")

	// ── run migration ────────────────────────────────────────────────────────

	require.NoError(t, nextupgrade.ReconstructTrancheKeys(ctx, app.AppCodec(), app.DexKeeper))

	// ── post-upgrade assertions ──────────────────────────────────────────────

	require.Len(t, app.DexKeeper.GetAllInactiveLimitOrderTranche(ctx), 3, "post-upgrade: entry count must not change")

	// --- tk-18498162 → tk-00000000000018498162 ---

	migratedKey1 := dextypes.NewTrancheKey(18498162)

	migratedTranche1, found := app.DexKeeper.GetInactiveLimitOrderTranche(ctx, &dextypes.LimitOrderTrancheKey{
		TradePairId:           pairID2,
		TrancheKey:            migratedKey1,
		TickIndexTakerToMaker: 67831,
	})
	require.True(t, found, "migrated tranche tk-18498162 must exist under new key")
	require.Equal(t, migratedKey1, migratedTranche1.Key.TrancheKey)

	_, found = app.DexKeeper.GetInactiveLimitOrderTranche(ctx, &dextypes.LimitOrderTrancheKey{
		TradePairId:           pairID2,
		TrancheKey:            "tk-18498162",
		TickIndexTakerToMaker: 67831,
	})
	require.False(t, found, "old key tk-18498162 must no longer exist")

	// --- tk-18291522 → tk-00000000000018291522 ---

	migratedKey2 := dextypes.NewTrancheKey(18291522)

	migratedTranche2, found := app.DexKeeper.GetInactiveLimitOrderTranche(ctx, &dextypes.LimitOrderTrancheKey{
		TradePairId:           pairID2,
		TrancheKey:            migratedKey2,
		TickIndexTakerToMaker: 67961,
	})
	require.True(t, found, "migrated tranche tk-18291522 must exist under new key")
	require.Equal(t, migratedKey2, migratedTranche2.Key.TrancheKey)

	_, found = app.DexKeeper.GetInactiveLimitOrderTranche(ctx, &dextypes.LimitOrderTrancheKey{
		TradePairId:           pairID2,
		TrancheKey:            "tk-18291522",
		TickIndexTakerToMaker: 67961,
	})
	require.False(t, found, "old key tk-18291522 must no longer exist")

	// --- 57m0a14awvr (base-36 key) must be unchanged ---

	untouched, found := app.DexKeeper.GetInactiveLimitOrderTranche(ctx, &dextypes.LimitOrderTrancheKey{
		TradePairId:           pairID1,
		TrancheKey:            "57m0a14awvr",
		TickIndexTakerToMaker: 0,
	})
	require.True(t, found, "base-36 tranche must still exist")
	require.Equal(t, "57m0a14awvr", untouched.Key.TrancheKey, "base-36 key must not be rewritten")
}

// TestReconstructLoTrancheUserLists verifies that LimitOrderTrancheUser entries stored under
// the old plain-decimal "tk-N" key are rewritten to the zero-padded "tk-%020d" format.
// Entries with the original base-36 sortable key (no "tk-" prefix) must remain unchanged.
func (suite *UpgradeTestSuite) TestReconstructLoTrancheUserLists() {
	app := suite.GetNeutronZoneApp(suite.ChainA)
	ctx := suite.ChainA.GetContext().WithChainID("neutron-1")
	t := suite.T()

	// ── pre-upgrade state ────────────────────────────────────────────────────

	pairID1 := dextypes.MustNewTradePairID(
		"ibc/773B4D0A3CD667B2275D5A4A7A2F0909C0BA0F4059C0B9181E680DDF4965DCC7",
		"ibc/B559A80D62249C8AA07A380E2A2BEA6E5CA9A6F079C912C3A9E9B494105E4F81",
	)
	// base-36 key — must NOT be migrated
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, &dextypes.LimitOrderTrancheUser{
		TradePairId:           pairID1,
		TickIndexTakerToMaker: -16365,
		TrancheKey:            "5atwxq41kck",
		Address:               "neutron12c20g3kvrvmqj3w5ep6vept6f77lunxyrrq44w",
		SharesOwned:           sdkmath.NewInt(5100000),
		SharesWithdrawn:       sdkmath.ZeroInt(),
		OrderType:             dextypes.LimitOrderType_GOOD_TIL_CANCELLED,
	})

	pairID2 := dextypes.MustNewTradePairID(
		"ibc/B559A80D62249C8AA07A380E2A2BEA6E5CA9A6F079C912C3A9E9B494105E4F81",
		"factory/neutron1r5qx58l3xx2y8gzjtkqjndjgx69mktmapl45vns0pa73z0zpn7fqgltnll/TAB",
	)
	// old plain-decimal key — must be migrated
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, &dextypes.LimitOrderTrancheUser{
		TradePairId:           pairID2,
		TickIndexTakerToMaker: -12041,
		TrancheKey:            "tk-3819855",
		Address:               "neutron12kmcmwljx7yplase4cjxqhry58fwvp5ljqu25a",
		SharesOwned:           sdkmath.NewInt(367803926),
		SharesWithdrawn:       sdkmath.ZeroInt(),
		OrderType:             dextypes.LimitOrderType_GOOD_TIL_CANCELLED,
	})
	app.DexKeeper.SetLimitOrderTrancheUser(ctx, &dextypes.LimitOrderTrancheUser{
		TradePairId:           pairID2,
		TickIndexTakerToMaker: -6931,
		TrancheKey:            "tk-4079303",
		Address:               "neutron12nrq3myjsfltjh5x8w8xcvxr8wpkef0vrpvxu4",
		SharesOwned:           sdkmath.NewInt(36000000),
		SharesWithdrawn:       sdkmath.ZeroInt(),
		OrderType:             dextypes.LimitOrderType_GOOD_TIL_CANCELLED,
	})

	require.Len(t, app.DexKeeper.GetAllLimitOrderTrancheUser(ctx), 3, "pre-upgrade: 3 tranche user entries")

	// ── run migration ────────────────────────────────────────────────────────

	require.NoError(t, nextupgrade.ReconstructTrancheKeys(ctx, app.AppCodec(), app.DexKeeper))

	// ── post-upgrade assertions ──────────────────────────────────────────────

	require.Len(t, app.DexKeeper.GetAllLimitOrderTrancheUser(ctx), 3, "post-upgrade: entry count must not change")

	// --- tk-3819855 → tk-00000000000003819855 ---

	migratedKey1 := dextypes.NewTrancheKey(3819855)

	migratedUser1, found := app.DexKeeper.GetLimitOrderTrancheUser(
		ctx,
		"neutron12kmcmwljx7yplase4cjxqhry58fwvp5ljqu25a",
		migratedKey1,
	)
	require.True(t, found, "migrated user tk-3819855 must exist under new key")
	require.Equal(t, migratedKey1, migratedUser1.TrancheKey)

	_, found = app.DexKeeper.GetLimitOrderTrancheUser(
		ctx,
		"neutron12kmcmwljx7yplase4cjxqhry58fwvp5ljqu25a",
		"tk-3819855",
	)
	require.False(t, found, "old key tk-3819855 must no longer exist")

	// --- tk-4079303 → tk-00000000000004079303 ---

	migratedKey2 := dextypes.NewTrancheKey(4079303)

	migratedUser2, found := app.DexKeeper.GetLimitOrderTrancheUser(
		ctx,
		"neutron12nrq3myjsfltjh5x8w8xcvxr8wpkef0vrpvxu4",
		migratedKey2,
	)
	require.True(t, found, "migrated user tk-4079303 must exist under new key")
	require.Equal(t, migratedKey2, migratedUser2.TrancheKey)

	_, found = app.DexKeeper.GetLimitOrderTrancheUser(
		ctx,
		"neutron12nrq3myjsfltjh5x8w8xcvxr8wpkef0vrpvxu4",
		"tk-4079303",
	)
	require.False(t, found, "old key tk-4079303 must no longer exist")

	// --- 5atwxq41kck (base-36 key) must be unchanged ---

	untouched, found := app.DexKeeper.GetLimitOrderTrancheUser(
		ctx,
		"neutron12c20g3kvrvmqj3w5ep6vept6f77lunxyrrq44w",
		"5atwxq41kck",
	)
	require.True(t, found, "base-36 tranche user must still exist")
	require.Equal(t, "5atwxq41kck", untouched.TrancheKey, "base-36 key must not be rewritten")
}
