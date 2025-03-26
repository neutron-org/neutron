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
	UpgradeName = "v6.0.0-rc0"

	// DropNtrnDenom is the denom of the Drop's NTRN token.
	DropNtrnDenom = "factory/neutron1ytalpjvxz7njekfep97sss2s83ezw6q8lt9spsvnd2d43ygys9gssy7ept/udntrn"
	// MainDAOContractAddress is the address of the Neutron DAO core contract.
	MainDAOContractAddress = "neutron1kvxlf27r0h7mzjqgdydqdf76dtlyvwz6u9q8tysfae53ajv8urtq4fdkvy"
	// VotingRegistryContractAddress is the address of the Neutron DAO voting registry contract.
	VotingRegistryContractAddress = "neutron1nusmqy8tmx5y2y5qrxprlm64fzwvjl9fhhn0qk5wy6mjkdrudsgqpmyywl"
)

// WARNING! Constants below represent tuples of addresses and code IDs of the new contracts. If you
// need to update an address, make sure to update the code ID as well (and vice versa).

const (
	DropCoreContractAddress = "neutron1wu9ng2pphg4g0a9d7ptq9ufqpcc7glhay33nhj79z4xs97qstj4q6un25a"
	DropCoreContractCodeID  = 11365 // TODO: populate when known

	StakingTrackerContractAddress = "neutron14lf29avfmv92cuj0avkp049d056d7mesw6g39p2nyfk9cg6c6shqzumtfj"
	StakingTrackerContractCodeID  = 11391 // TODO: populate when known

	StakingVaultContractAddress = "neutron12u7e2782xp0dqdlfqgpq2yy8k42gkmwpjvq3eehssvk7hsdkt5gq67fcrk"
	StakingVaultContractCodeID  = 11392 // TODO: populate when known

	StakingRewardsContractAddress = "neutron1h62p45vv3fg2q6sm00r93gqgmhqt9tfgq5hz33qyrhq8f0pqqj0s36wgc3"
	StakingRewardsContractCodeID  = 11393 // TODO: populate when known

	StakingInfoProxyContractAddress = "neutron1yrnnq3q3k74e2cllrapzc4rdgeg8y44wccurmnz87c0mhprr6ews7k5fz3"
	StakingInfoProxyContractCodeID  = 11394 // TODO: populate when known
)

const (
	// Constants defining the parameters of the assets redistribution process held by DAO.
	StakeWithValenceAmount = 50_000_000_000_000
	StakeWithDropAmount    = 50_000_000_000_000

	ValenceStaker = "neutron137kcd226g24frg3pczal4ux72k6lrk5pnfl482zceceyyldzjmqsrmgsrv"

	UsdcLpReceiver = "neutron1qaaf9lv99pwyeaf7ktw37wetyzpper28j6cltqgw600g3gtsac4s64wtx2"
	// USDC-NTRN
	UsdcLpDenom = "factory/neutron16puus9vjwq4xq0pkl59x30qwn5t48t7r90zqcgc5g8qsyu0u0fnskraxld/astroport/share" // TODO: populate when known

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
