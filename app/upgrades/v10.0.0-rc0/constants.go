package v10_0_0_rc0

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/neutron-org/neutron/v10/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v10.0.0-rc0"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{},
		Deleted: []string{
			"capability",
		},
	},
}
