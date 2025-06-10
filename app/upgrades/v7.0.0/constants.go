package v700

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/neutron-org/neutron/v7/app/upgrades"
	stateverifier "github.com/neutron-org/neutron/v7/x/state-verifier/types"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v7.0.0"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			stateverifier.ModuleName,
		},
		// to avoid import of ccconsumer and auction modules just for their removal, i use just string here
		Deleted: []string{
			"ccvconsumer",
			"auction",
		},
	},
}
