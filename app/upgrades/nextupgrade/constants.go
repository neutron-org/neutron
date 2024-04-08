package nextupgrade

import (
	"cosmossdk.io/store/types"

	"github.com/neutron-org/neutron/v3/app/upgrades"
	globalfeetypes "github.com/neutron-org/neutron/v3/x/globalfee/types"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "NextUpgrade"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: types.StoreUpgrades{
		Added: []string{
			globalfeetypes.ModuleName,
		},
	},
}
