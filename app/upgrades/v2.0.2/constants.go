package v202

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v2/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v2.0.2"

	AuctionParamsMaxBundleSize          = 2
	AuctionParamsFrontRunningProtection = true
)

var (
	Upgrade = upgrades.Upgrade{
		UpgradeName:          UpgradeName,
		CreateUpgradeHandler: CreateUpgradeHandler,
	}
	AuctionParamsReserveFee      = sdk.Coin{Denom: "untrn", Amount: sdk.NewInt(1_000_000)}
	AuctionParamsMinBidIncrement = sdk.Coin{Denom: "untrn", Amount: sdk.NewInt(1_000_000)}
	AuctionParamsProposerFee     = math.LegacyNewDecWithPrec(25, 2)
)
