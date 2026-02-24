package nextupgrade

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/neutron-org/neutron/v10/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "nextupgrade"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			"gov",
			"mint",
			"distribution",
		},
		Deleted: []string{
			"adminmodule",
			"harpoon",
			"revenue",
			"feeburner",
		},
	},
}
