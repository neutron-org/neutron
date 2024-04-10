package nextupgrade

import (
	storetypes "cosmossdk.io/store/types"
	marketmaptypes "github.com/skip-mev/slinky/x/marketmap/types"
	oracletypes "github.com/skip-mev/slinky/x/oracle/types"

	"github.com/neutron-org/neutron/v3/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName        = "nextupgrade"
	AdminModuleAddress = "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{marketmaptypes.ModuleName, oracletypes.ModuleName},
	},
}
