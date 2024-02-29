package v300

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	auctionkeeper "github.com/skip-mev/block-sdk/x/auction/keeper"
	auctiontypes "github.com/skip-mev/block-sdk/x/auction/types"

	"github.com/neutron-org/neutron/v3/app/upgrades"

	feeburnerkeeper "github.com/neutron-org/neutron/v3/x/feeburner/keeper"

	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.UpgradeKeepers,
	_ upgrades.StoreKeys,
	_ codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Setting block-sdk params...")
		err = setAuctionParams(ctx, keepers.FeeBurnerKeeper, keepers.AuctionKeeper)
		if err != nil {
			return nil, err
		}

		if err := migrateICSOutstandingDowntime(ctx, keepers); err != nil {
			return vm, fmt.Errorf("failed to migrate ICS outstanding downtime: %w", err)
		}

		recalculateSlashingMissedBlocksCounter(ctx, keepers)

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

func setAuctionParams(ctx sdk.Context, feeBurnerKeeper *feeburnerkeeper.Keeper, auctionKeeper auctionkeeper.Keeper) error {
	treasury := feeBurnerKeeper.GetParams(ctx).TreasuryAddress
	_, data, err := bech32.DecodeAndConvert(treasury)
	if err != nil {
		return err
	}

	auctionParams := auctiontypes.Params{
		MaxBundleSize:          AuctionParamsMaxBundleSize,
		EscrowAccountAddress:   data,
		ReserveFee:             AuctionParamsReserveFee,
		MinBidIncrement:        AuctionParamsMinBidIncrement,
		FrontRunningProtection: AuctionParamsFrontRunningProtection,
		ProposerFee:            AuctionParamsProposerFee,
	}
	return auctionKeeper.SetParams(ctx, auctionParams)
}

// Sometime long ago we decreased SlashWindow to 36k on pion-1 testnet (the param is untouched on neutron-1 mainnet),
// from that time MissedBlockCounter is wrong
// We need to set to a proper value.
// Proper value is: MissedBlocksCounter = sum_of_missed_blocks_in_bitmap(bitmap).
// Since the param is untouched on neutron-1 mainnet, this method does not change anything during the migration on mainnet.
func recalculateSlashingMissedBlocksCounter(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) {
	ctx.Logger().Info("Starting recalculating MissedBlocksCounter for validators...")
	signingInfos := make([]slashingtypes.ValidatorSigningInfo, 0)
	consAddresses := make([]sdk.ConsAddress, 0)

	keepers.SlashingKeeper.IterateValidatorSigningInfos(ctx, func(addr sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) bool {
		signingInfos = append(signingInfos, info)
		consAddresses = append(consAddresses, addr)
		return false
	})

	for i, info := range signingInfos {
		ctx.Logger().Info("MissedBlocks recalculating", "Validator", info.Address, "old MissedBlocksCounter", info.MissedBlocksCounter)

		missedBlocksForValidator := int64(0)

		keepers.SlashingKeeper.IterateValidatorMissedBlockBitArray(ctx, consAddresses[i], func(index int64, missed bool) bool {
			if missed {
				missedBlocksForValidator++
			}
			return false
		})

		info.MissedBlocksCounter = missedBlocksForValidator

		ctx.Logger().Info("MissedBlocks recalculating", "Validator", info.Address, "new MissedBlocksCounter", info.MissedBlocksCounter)

		keepers.SlashingKeeper.SetValidatorSigningInfo(ctx, consAddresses[i], info)
	}

	ctx.Logger().Info("Finished recalculating MissedBlocksCounter for validators")
}

func migrateICSOutstandingDowntime(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) error {
	ctx.Logger().Info("Migrating ICS outstanding downtime...")

	downtimes := keepers.ConsumerKeeper.GetAllOutstandingDowntimes(ctx)
	for _, od := range downtimes {
		consAddr, err := sdk.ConsAddressFromBech32(od.ValidatorConsensusAddress)
		if err != nil {
			return err
		}
		keepers.ConsumerKeeper.DeleteOutstandingDowntime(ctx, consAddr)
	}

	ctx.Logger().Info("Finished ICS outstanding downtime")

	return nil
}
