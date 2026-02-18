package v900

import (
	storetypes "cosmossdk.io/store/types"

	coinfactorytypes "github.com/neutron-org/neutron/v10/x/coinfactory/types"

	"github.com/neutron-org/neutron/v10/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v9.0.0"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{coinfactorytypes.StoreKey},
	},
}
