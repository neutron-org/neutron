package v3

import (
	ccvprovider "github.com/cosmos/interchain-security/x/ccv/provider/types"

	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/neutron-org/neutron/app/upgrades"
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
