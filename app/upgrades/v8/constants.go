package v8

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/neutron-org/neutron/v7/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v8"
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
