package v600

import (
	storetypes "cosmossdk.io/store/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	revenuetypes "github.com/neutron-org/neutron/v6/x/revenue/types"

	harpoontypes "github.com/neutron-org/neutron/v6/x/harpoon/types"

	"github.com/neutron-org/neutron/v6/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "sovereign"

	// DropNtrnDenom is the denom of the Drop's NTRN token.
	DropNtrnDenom = "factory/neutron1lzfk4aj26jz7gd3c4umxah9d22ezy8xfql677kev37vd0mq8y3tsn78saz/udntrn"
	// MainDAOContractAddress is the address of the Neutron DAO core contract.
	MainDAOContractAddress = "neutron1yw4xvtc43me9scqfr2jr2gzvcxd3a9y4eq7gaukreugw2yd2f8ts8g30fq"
	// VotingRegistryContractAddress is the address of the Neutron DAO voting registry contract.
	VotingRegistryContractAddress = "neutron13ehuhysn5mqjeaheeuew2gjs785f6k7jm8vfsqg3jhtpkwppcmzqu8vxt2"
)

// WARNING! Constants below represent tuples of addresses and code IDs of the new contracts. If you
// need to update an address, make sure to update the code ID as well (and vice versa).

const (
	DropCoreContractAddress = "neutron18ecx6f2ywwnfxsql2l98jscw97lezczx8ax0g5wp8uj9rm95m0ls798cdq"
	DropCoreContractCodeID  = 25 // TODO: populate when known

	StakingTrackerContractAddress = "neutron1nyuryl5u5z04dx4zsqgvsuw7fe8gl2f77yufynauuhklnnmnjncqcls0tj"
	StakingTrackerContractCodeID  = 20 // TODO: populate when known

	StakingVaultContractAddress = "neutron1jarq7kgdyd7dcfu2ezeqvg4w4hqdt3m5lv364d8mztnp9pzmwwwqjw7fvg"
	StakingVaultContractCodeID  = 21 // TODO: populate when known

	StakingRewardsContractAddress = "neutron1mygmlglvg9w45n3s6m6d4txneantmupy0sa0vy63angpvj0qp7usep7kdz"
	StakingRewardsContractCodeID  = 22 // TODO: populate when known

	StakingInfoProxyContractAddress = "neutron1xx35wwa2nhfvfm50lj3ukv077mjxuy9pefxxnctxe9kczk6tz3hqpxknre"
	StakingInfoProxyContractCodeID  = 23 // TODO: populate when known
)

const (
	// Constants defining the parameters of the assets redistribution process held by DAO.
	TotalToStakeAmount     = int64(50_000_000_000_000)                   // untrn
	StakeWithValenceAmount = 45_000_000_000_000                          // 90% of TotalToStakeAmount
	StakeWithDropAmount    = TotalToStakeAmount - StakeWithValenceAmount // 10% of TotalToStakeAmount

	ValenceStaker = "neutron1yvxarc3r8agzzky6g4zdxhk5xc59j7rdw2pugjarwskt8jmpkgus9jqvwk"

	UsdcLpReceiver = "neutron1l3pk6xsc8p74gwrduxj5qp7djqyx9j5uweufrw5kp8d37xc3hjqs53euye"
	UsdcLpDenom    = "factory/neutron1czkddm6xqyfa6ukzxqmf65tl4tudry4kve0n8fs5yfc8g6zv52lqznmnnl/astroport/share" // TODO: populate when known

	dntrnNtrnLiqAmount   = int64(22_500_000_000_000)
	dntrnNtrnLiqProvider = "neutron1dmhxfvggstv2k9xd4rg2nmvsdwesh9mj20dl70lmcala93zmfnfsmgls2c"

	RewardContract = 6_000_000_000_000
	RevenueModule  = 1_000_000_000_000
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
