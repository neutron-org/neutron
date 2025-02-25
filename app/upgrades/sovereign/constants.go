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
	UpgradeName = "sovereign3"

	StakingTrackerContractAddress = "neutron1r50m5lafnmctat4xpvwdpzqndynlxt2skhr4fhzh76u0qar2y9hqyaf9ku"
	StakingVaultContractAddress   = "neutron1xqju350rsytdrgvyc27zljjvpygefmnvdjnfel8rcannyn6zt2rq8fjxj2"
	StakingRewardsContractAddress = ""
	VotingRegistryContractAddress = "neutron13ehuhysn5mqjeaheeuew2gjs785f6k7jm8vfsqg3jhtpkwppcmzqu8vxt2"
	MainDAOContractAddress        = "neutron1yw4xvtc43me9scqfr2jr2gzvcxd3a9y4eq7gaukreugw2yd2f8ts8g30fq"
	DropDelegateContract          = "neutron1ypj29p92305sc8r9azz9h8jkjte8r0xx5xnw6lgdcezvnc777twswxgp99"
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

// CodesToPin - codes to pin of the new stored contracts:
// rewards, vault, tracker, proxy
var CodesToPin = []uint64{
	0,
}
