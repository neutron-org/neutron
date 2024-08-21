package v421testnet

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/neutron-org/neutron/v4/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v4.2.1-testnet"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{},
	},
}
