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
	DropNtrnDenom = "factory/neutron1jz90vam2a4glwll770psh5tyg72k0kcvwtfrx4ysx2mac9ynv8rq0uevh9/udntrn"
	// MainDAOContractAddress is the address of the Neutron DAO core contract.
	MainDAOContractAddress = "neutron1suhgf5svhu4usrurvxzlgn54ksxmn8gljarjtxqnapv8kjnp4nrstdxvff"
	// VotingRegistryContractAddress is the address of the Neutron DAO voting registry contract.
	VotingRegistryContractAddress = "neutron1f6jlx7d9y408tlzue7r2qcf79plp549n30yzqjajjud8vm7m4vdspg933s"
)

// WARNING! Constants below represent tuples of addresses and code IDs of the new contracts. If you
// need to update an address, make sure to update the code ID as well (and vice versa).

const (
	DropCoreContractAddress = "neutron1l8s4xge4s0hkvd8y7a8tkejakjmdcg0mhpst5sdufwfex2luhgrsugu0h8"
	DropCoreContractCodeID  = 3237 // TODO: populate when known

	StakingTrackerContractAddress = "neutron1hja44wkskcmvc36xhp5axlcflzvlw9mzq98pyksa3ftvrx83hceqg4z45f"
	StakingTrackerContractCodeID  = 3248 // TODO: populate when known

	StakingVaultContractAddress = "neutron1krrz6l5gurrmxnx2903mnx48s5ge9chepegnwtrhuzt4r34udy9q87q5th"
	StakingVaultContractCodeID  = 3249 // TODO: populate when known

	StakingRewardsContractAddress = "neutron1w04ykk9s8mmpjcakvfuetugg2uyqc6t3mcr8j5cqx3tv08j7lvcqtf3uth"
	StakingRewardsContractCodeID  = 3250 // TODO: populate when known

	StakingInfoProxyContractAddress = "neutron1w7y9gwhdtgs3kv2dnl5gpaxzp6k720uekjhgx2qtx9qv9nkadn0sxpcrqt"
	StakingInfoProxyContractCodeID  = 3251 // TODO: populate when known
)

const (
	// Constants defining the parameters of the assets redistribution process held by DAO.
	TotalToStakeAmount     = int64(250_000_000_000_000)                  // untrn
	StakeWithValenceAmount = 240_000_000_000_000                         // 90% of TotalToStakeAmount
	StakeWithDropAmount    = TotalToStakeAmount - StakeWithValenceAmount // 10% of TotalToStakeAmount

	ValenceStaker = "neutron1yvxarc3r8agzzky6g4zdxhk5xc59j7rdw2pugjarwskt8jmpkgus9jqvwk"

	USDC_LP_Receiver = "neutron1l3pk6xsc8p74gwrduxj5qp7djqyx9j5uweufrw5kp8d37xc3hjqs53euye"
	// USDC-NTRN
	USDC_LP_Denom = "factory/neutron145z3nj7yqft2vpugr5a5p7jsnagvms90tvtej45g4s0xkqalhy7sj20vgz/astroport/share" // TODO: populate when known

	dNTRN_NTRN_LiqAmount   = int64(25_000_000_000_000)
	dNTRN_NTRN_LiqProvider = "neutron1dmhxfvggstv2k9xd4rg2nmvsdwesh9mj20dl70lmcala93zmfnfsmgls2c"

	RewardContract = 6_000_000_000_000
	RevenueModule  = 6_000_000_000_000
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
