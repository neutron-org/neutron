package v800_rc0

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/neutron-org/neutron/v8/app/upgrades"
	stateverifier "github.com/neutron-org/neutron/v8/x/state-verifier/types"
)

// This v8.0.0-rc0 upgrade is only for pion-1 chain, since it was already migrated to v7.0.0, thus
// v8.0.0-rc0 contains only migration from v7.0.0 to v8.0.0

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v8.0.0-rc0"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			stateverifier.ModuleName,
		},
		Deleted: []string{},
	},
}
