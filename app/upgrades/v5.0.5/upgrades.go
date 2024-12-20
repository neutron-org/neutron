package v505

import (
	"context"
	"fmt"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/neutron-org/neutron/v5/app/upgrades"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.UpgradeKeepers,
	_ upgrades.StoreKeys,
	_ codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)

		ctx.Logger().Info("Starting module migrations...")

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		transferChannels := keepers.ChannelKeeper.GetAllChannelsWithPortPrefix(ctx, keepers.TransferKeeper.GetPort(ctx))
		for _, channel := range transferChannels {
			escrowAddress := transfertypes.GetEscrowAddress(channel.PortId, channel.ChannelId)
			ctx.Logger().Info("Saving escrow address", "port_id", channel.PortId, "channel_id",
				channel.ChannelId, "address", escrowAddress.String())
			keepers.TokenFactoryKeeper.StoreEscrowAddress(ctx, escrowAddress)
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}
