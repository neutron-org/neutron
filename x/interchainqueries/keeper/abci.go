package keeper

import (
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/x/interchainqueries/types"
)

// EndBlocker of interchainquery module
func (k Keeper) EndBlocker(ctx sdk.Context) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	events := sdk.Events{}

	// emit events for periodic queries
	k.IterateRegisteredQueries(ctx, func(_ int64, registeredQuery types.RegisteredQuery) (stop bool) {
		if uint64(ctx.BlockHeight()) >= registeredQuery.LastEmittedHeight+registeredQuery.UpdatePeriod {
			k.Logger(ctx).Debug("EndBlocker: interchainquery event emitted", "id", registeredQuery.Id)
			event := sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueQuery),
				sdk.NewAttribute(types.AttributeKeyQueryID, strconv.FormatUint(registeredQuery.Id, 10)),
				sdk.NewAttribute(types.AttributeKeyOwner, registeredQuery.Owner),
				sdk.NewAttribute(types.AttributeKeyConnectionID, registeredQuery.ConnectionId),
				sdk.NewAttribute(types.AttributeKeyQueryType, registeredQuery.QueryType),
				sdk.NewAttribute(types.AttributeTransactionsFilterQuery, registeredQuery.TransactionsFilter),
				sdk.NewAttribute(types.AttributeKeyKVQuery, types.KVKeys(registeredQuery.Keys).String()),
			)
			events = append(events, event)
			registeredQuery.LastEmittedHeight = uint64(ctx.BlockHeight())
			if err := k.SaveQuery(ctx, registeredQuery); err != nil {
				k.Logger(ctx).Error("EndBlocker: failed to save query", "error", err)
			}
			k.Logger(ctx).Debug("EndBlocker: event successfully added to events list", "id", registeredQuery)
		}
		return false
	})

	if len(events) > 0 {
		ctx.EventManager().EmitEvents(events)
		k.Logger(ctx).Debug("EndBlocker: processed events", "events", events)
	}
}
