package v600

import (
	storetypes "cosmossdk.io/store/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	revenuetypes "github.com/neutron-org/neutron/v7/x/revenue/types"

	harpoontypes "github.com/neutron-org/neutron/v7/x/harpoon/types"

	"github.com/neutron-org/neutron/v7/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v6.0.0"

	// DropNtrnDenom is the denom of the Drop's NTRN token.
	DropNtrnDenom = "factory/neutron1frc0p5czd9uaaymdkug2njz7dc7j65jxukp9apmt9260a8egujkspms2t2/udntrn"
	// MainDAOContractAddress is the address of the Neutron DAO core contract.
	MainDAOContractAddress = "neutron1suhgf5svhu4usrurvxzlgn54ksxmn8gljarjtxqnapv8kjnp4nrstdxvff"
	// VotingRegistryContractAddress is the address of the Neutron DAO voting registry contract.
	VotingRegistryContractAddress = "neutron1f6jlx7d9y408tlzue7r2qcf79plp549n30yzqjajjud8vm7m4vdspg933s"
)

// WARNING! Constants below represent tuples of addresses and code IDs of the new contracts. If you
// need to update an address, make sure to update the code ID as well (and vice versa).

const (
	DropPuppeteerContractAddress = "neutron17jsl4t4hhaw37tnhenskrfntm7mv44wzjr3f990hx4p9r5m0gzdqquhtd3"
	DropCoreContractAddress      = "neutron1lsxvdyvmexak084wdty2yvsq5gj3wt7wm4jaw34yseat7r4qjffqlxlcua"
	DropCoreContractCodeID       = 3332

	StakingTrackerContractAddress = "neutron1kf9yq7vuyj9rshwpr52xru779y832g7jpgyysprvpm9xzu2m6mlsm6r64n"
	StakingTrackerContractCodeID  = 3345

	StakingVaultContractAddress = "neutron19j2m9enzvq4kpd72tr3cz46z2kq6rnedc2q4pj6w5wq6v86va58qkled36"
	StakingVaultContractCodeID  = 3346

	StakingRewardsContractAddress = "neutron1gqq3c735pj6ese3yru5xr6ud0fvxgltxesygvyyzpsrt74v6yg4sgkrgwq"
	StakingRewardsContractCodeID  = 3347

	StakingInfoProxyContractAddress = "neutron187ufx8wf0neqduppvkn0rknpg8j2s59qdhua80vgfhl7yclk8rrsnjukgq"
	StakingInfoProxyContractCodeID  = 3348
)

const (
	// Constants defining the parameters of the assets redistribution process held by DAO.
	StakeWithValenceAmount = 150_000_000_000_000
	StakeWithDropAmount    = 100_000_000_000_000

	ValenceStaker = "neutron1d846g3px2u0dhs6r7k46e49nhgwcxzzw9973rau07lp9cpsj0v3q05ahf4"

	UsdcLpReceiver = "neutron1d58c25fw3hwpjvg9dzgr2m235qpgtsc7stjt7u08kqg8jd583fgsyr5ytg"
	UsdcLpDenom    = "factory/neutron18c8qejysp4hgcfuxdpj4wf29mevzwllz5yh8uayjxamwtrs0n9fshq9vtv/astroport/share"

	RewardContract = 6_000_000_000_000
	RevenueModule  = 700_000_000_000
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
