package v400

import (
	storetypes "cosmossdk.io/store/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/neutron-org/neutron/v4/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "NextUpgrade"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			stakingtypes.ModuleName,
		},
	},
}
