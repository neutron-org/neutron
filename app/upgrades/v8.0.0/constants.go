package v800

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/neutron-org/neutron/v8/app/upgrades"
	stateverifier "github.com/neutron-org/neutron/v8/x/state-verifier/types"
)

// This v8.0.0 upgrade is meant only for the neutron-1 chain (mainnet),
// since it was never upgrade to v7 release, but we still need to do necessary migrations,
// thus this upgrade contains migrations from v6 to v7 and also from v7 to v8

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v8.0.0"
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
