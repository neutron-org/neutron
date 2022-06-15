package keeper

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lidofinance/gaia-wasm-zone/x/interchainqueries/types"
)

// EndBlocker of interchainquery module
func (k Keeper) EndBlocker(ctx sdk.Context) {
	events := sdk.Events{}

	// emit events for periodic queries
	k.IterateRegisteredQueries(ctx, func(_ int64, registeredQuery types.RegisteredQuery) (stop bool) {
		if registeredQuery.LastEmittedHeight+registeredQuery.UpdatePeriod == uint64(ctx.BlockHeight()) {
			k.Logger(ctx).Info("Interchainquery event emitted", "id", registeredQuery.Id)
			event := sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueQuery),
				sdk.NewAttribute(types.AttributeKeyQueryID, strconv.FormatUint(registeredQuery.Id, 10)),
				sdk.NewAttribute(types.AttributeKeyZoneID, registeredQuery.ZoneId),
				sdk.NewAttribute(types.AttributeQueryType, registeredQuery.QueryType),
				sdk.NewAttribute(types.AttributeQueryData, registeredQuery.QueryData),
			)

			events = append(events, event)
			registeredQuery.LastEmittedHeight = uint64(ctx.BlockHeight())
			if err := k.SaveQuery(ctx, registeredQuery); err != nil {
				k.Logger(ctx).Error("failed to save query: %v", err)
			}

		}
		return false
	})

	if len(events) > 0 {
		ctx.EventManager().EmitEvents(events)
	}
}
