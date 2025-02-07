package v510

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/neutron-org/neutron/v5/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName                = "v5.1.0"
	MarketMapAuthorityMultisig = "neutron1ldvrvhyvtssm0ptdq23hhaltprx8ctmjh92kzfs55sz2z997n76s72wr86"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades:        storetypes.StoreUpgrades{},
}
