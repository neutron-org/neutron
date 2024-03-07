package nextupgrade

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	storetypes "cosmossdk.io/store/types"
	auctiontypes "github.com/skip-mev/block-sdk/v2/x/auction/types"

	"github.com/neutron-org/neutron/v2/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "nextupgrade"

	AuctionParamsMaxBundleSize          = 2
	AuctionParamsFrontRunningProtection = true
)

var (
	Upgrade = upgrades.Upgrade{
		UpgradeName:          UpgradeName,
		CreateUpgradeHandler: CreateUpgradeHandler,
		StoreUpgrades: storetypes.StoreUpgrades{
			Added: []string{auctiontypes.ModuleName},
		},
	}
	AuctionParamsReserveFee      = sdk.Coin{Denom: "untrn", Amount: math.NewInt(500_000)}
	AuctionParamsMinBidIncrement = sdk.Coin{Denom: "untrn", Amount: math.NewInt(100_000)}
	AuctionParamsProposerFee     = math.LegacyNewDecWithPrec(25, 2)
)
