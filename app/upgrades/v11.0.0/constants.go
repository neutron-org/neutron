package v11_0_0

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/neutron-org/neutron/v11/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v11.0.0"
)

var Deleted = []string{
	"adminmodule",
	"harpoon",
	"revenue",
	"feeburner",
}

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			"gov",
			"mint",
			"distribution",
		},
		Deleted: Deleted,
	},
}
