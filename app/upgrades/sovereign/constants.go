package v600

import (
	storetypes "cosmossdk.io/store/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"

	harpoontypes "github.com/neutron-org/neutron/v5/x/harpoon/types"

	"github.com/neutron-org/neutron/v5/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "sovereign"

	// DropNtrnDenom is the denom of the Drop's NTRN token.
	DropNtrnDenom = "TODO: populate when known"
	// MainDAOContractAddress is the address of the Neutron DAO core contract.
	MainDAOContractAddress = "neutron1suhgf5svhu4usrurvxzlgn54ksxmn8gljarjtxqnapv8kjnp4nrstdxvff"
	// VotingRegistryContractAddress is the address of the Neutron DAO voting registry contract.
	VotingRegistryContractAddress = "neutron1f6jlx7d9y408tlzue7r2qcf79plp549n30yzqjajjud8vm7m4vdspg933s"
)

// WARNING! Constants below represent tuples of addresses and code IDs of the new contracts. If you
// need to update an address, make sure to update the code ID as well (and vice versa).

const (
	DropCoreContractAddress = "TODO: populate when known"
	DropCoreContractCodeID  = 0 // TODO: populate when known

	StakingTrackerContractAddress = "TODO: populate when known"
	StakingTrackerContractCodeID  = 0 // TODO: populate when known

	StakingVaultContractAddress = "TODO: populate when known"
	StakingVaultContractCodeID  = 0 // TODO: populate when known

	StakingRewardsContractAddress = "TODO: populate when known"
	StakingRewardsContractCodeID  = 0 // TODO: populate when known

	StakingInfoProxyContractAddress = "TODO: populate when known"
	StakingInfoProxyContractCodeID  = 0 // TODO: populate when known
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			stakingtypes.ModuleName,
			harpoontypes.ModuleName,
			revenuetypes.ModuleName,
		},
	},
}

// CodesToPin is a list of code IDs to pin. Contains code IDs of contracts related to the Neutron
// governance, staking, and DeFi.
var CodesToPin = []uint64{
	DropCoreContractCodeID,
	StakingTrackerContractCodeID,
	StakingVaultContractCodeID,
	StakingRewardsContractCodeID,
	StakingInfoProxyContractCodeID,
}
