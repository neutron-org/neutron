package v400

import (
	storetypes "cosmossdk.io/store/types"
	marketmaptypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	oracletypes "github.com/skip-mev/connect/v2/x/oracle/types"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"

	dynamicfeestypes "github.com/neutron-org/neutron/v4/x/dynamicfees/types"

	"github.com/neutron-org/neutron/v4/app/upgrades"
	globalfeetypes "github.com/neutron-org/neutron/v4/x/globalfee/types"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v4.0.1"

	// MarketMapAuthorityMultisig defines the address of a market-map authority governed by a
	// multi-sig of contributors.
	MarketMapAuthorityMultisig = "neutron1ua63s43u2p4v38pxhcxmps0tj2gudyw2hfeetz"

	DecimalsAdjustment = 1_000_000_000_000
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			globalfeetypes.ModuleName,
			marketmaptypes.ModuleName,
			oracletypes.ModuleName,
			feemarkettypes.ModuleName,
			dynamicfeestypes.ModuleName,
		},
	},
}
