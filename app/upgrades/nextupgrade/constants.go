package nextupgrade

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/neutron-org/neutron/v8/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "nextupgrade"
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
