package keeper

import (
	"fmt"
	"strconv"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v3/x/dex/types"
)

var (
	AttributeWithdrawn                 = "total_withdrawn"
	AttributeGasConsumed               = "gas_consumed"
	MetricLabelTotalOrdersExpired      = "total_orders_expired"
	MetricLabelTotalLimitOrders        = "total_orders_limit"
	MetricLabelTotalTickLiquiditiesInc = "total_tick_liqidities_inc"
	MetricLabelTotalTickLiquiditiesDec = "total_tick_liqidities_dec"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		bankKeeper types.BankKeeper
		authority  string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		bankKeeper: bankKeeper,
		authority:  authority,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func getEventsWithdrawnAmount(coins sdk.Coins) sdk.Events {
	events := sdk.Events{}
	for _, coin := range coins {
		event := sdk.NewEvent(
			types.EventTypeNeutronMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeDenom, coin.Denom),
			sdk.NewAttribute(types.AttributeWithdrawn, coin.Amount.String()),
		)
		events = append(events, event)
	}
	return events
}

func getEventsGasConsumed(gasBefore, gasAfter sdk.Gas) sdk.Events {
	return sdk.Events{
		sdk.NewEvent(
			types.EventTypeNeutronMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeGasConsumed, strconv.FormatUint(gasAfter-gasBefore, 10)),
		),
	}
}

func getEventsIncExpiredOrders() sdk.Events {
	return sdk.Events{
		sdk.NewEvent(
			types.EventTypeNeutronMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeIncExpiredOrders),
		),
	}
}

func getEventsDecExpiredOrders() sdk.Events {
	return sdk.Events{
		sdk.NewEvent(
			types.EventTypeNeutronMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeDecExpiredOrders),
		),
	}
}

func getEventsIncTotalOrders() sdk.Events {
	return sdk.Events{
		sdk.NewEvent(
			types.EventTypeNeutronMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeTotalLimitOrders),
		),
	}
}

func getEventsIncTotalTickLiquidities() sdk.Events {
	return sdk.Events{
		sdk.NewEvent(
			types.EventTypeNeutronMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeTotalTickLiquiditiesInc),
		),
	}
}

func getEventsDecTotalTickLiquidities() sdk.Events {
	return sdk.Events{
		sdk.NewEvent(
			types.EventTypeNeutronMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeTotalTickLiquiditiesDec),
		),
	}
}
