package v400

import (
	storetypes "cosmossdk.io/store/types"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
	marketmaptypes "github.com/skip-mev/slinky/x/marketmap/types"
	oracletypes "github.com/skip-mev/slinky/x/oracle/types"

	dynamicfeestypes "github.com/neutron-org/neutron/v4/x/dynamicfees/types"

	"github.com/neutron-org/neutron/v4/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v4.0.0"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			marketmaptypes.ModuleName,
			oracletypes.ModuleName,
			feemarkettypes.ModuleName,
			dynamicfeestypes.ModuleName,
		},
	},
}
