package v030

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	ccvprovider "github.com/cosmos/interchain-security/v4/x/ccv/provider/types"

	"github.com/neutron-org/neutron/v2/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrades name.
	UpgradeName = "v0.3.0"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			ccvprovider.ModuleName,
		},
	},
}
