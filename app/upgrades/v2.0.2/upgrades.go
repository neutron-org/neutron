package v202

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/neutron-org/neutron/v2/app/upgrades"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	auctionkeeper "github.com/skip-mev/block-sdk/x/auction/keeper"
	auctiontypes "github.com/skip-mev/block-sdk/x/auction/types"
	feeburnerkeeper "github.com/neutron-org/neutron/v2/x/feeburner/keeper"
)

func CreateUpgradeHandler(
	_ *module.Manager,
	_ module.Configurator,
	keepers *upgrades.UpgradeKeepers,
	_ upgrades.StoreKeys,
	_ codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Setting block-sdk params...")
		err := setAuctionParams(ctx, keepers.FeeBurnerKeeper, keepers.AuctionKeeper)
		if err != nil {
			return nil, err
		}

		return vm, nil
	}
}

func setAuctionParams(ctx sdk.Context, feeBurnerKeeper *feeburnerkeeper.Keeper, auctionKeeper auctionkeeper.Keeper) error {
	treasury := feeBurnerKeeper.GetParams(ctx).TreasuryAddress
	_, data, err := bech32.DecodeAndConvert(treasury)
	if err != nil {
		return err
	}
	fmt.Println("treasury addy", feeBurnerKeeper.GetParams(ctx).TreasuryAddress)

	auctionParams := auctiontypes.Params{
		MaxBundleSize:          AuctionParamsMaxBundleSize,
		EscrowAccountAddress:   data,
		ReserveFee: AuctionParamsReserveFee,
		MinBidIncrement:        AuctionParamsMinBidIncrement,
		FrontRunningProtection: AuctionParamsFrontRunningProtection,
		ProposerFee:            AuctionParamsProposerFee,
	}
	return auctionKeeper.SetParams(ctx, auctionParams)
}
