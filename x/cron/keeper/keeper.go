package keeper

import (
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/neutron-org/neutron/x/cron/types"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace
		permKeeper *wasmkeeper.PermissionedKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	permKeeper *wasmkeeper.PermissionedKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
		permKeeper: permKeeper,
	}
}

// TODO: DOC comments

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k *Keeper) CheckTimer(ctx sdk.Context) {
	schedules := k.getSchedulesReadyForExecution(ctx)

	for _, schedule := range schedules {
		wasmMsgServer := wasmkeeper.NewMsgServerImpl(k.permKeeper)
		k.executeSchedule(ctx, wasmMsgServer, schedule)
	}
}

// period in blocks
func (k *Keeper) AddSchedule(ctx sdk.Context, contractAddr sdk.AccAddress, name string, period uint64, msgs []wasmtypes.MsgExecuteContract) {
	// TODO: check contractAddr is DAO admin
	schedule := types.Schedule{
		Name:              name,
		Period:            period,
		Msgs:              msgs,
		LastExecuteHeight: uint64(ctx.BlockHeight()), // lets execute newly added schedule on `now + period` block // TODO: ok conversion?
	}
	k.storeSchedule(ctx, schedule)
}

func (k *Keeper) RemoveSchedule(ctx sdk.Context, contractAddr sdk.AccAddress, name string) {
	// TODO: check contractAddr is DAO admin or Security DAO admin
	k.removeSchedule(ctx, name)
}

func (k *Keeper) GetAllSchedules(ctx sdk.Context) []types.Schedule {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ScheduleKey)

	res := make([]types.Schedule, 0)

	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var schedule types.Schedule
		k.cdc.MustUnmarshal(iterator.Value(), &schedule)
		res = append(res, schedule)
	}

	return res
}

func (k *Keeper) GetSchedule(ctx sdk.Context, name string) (*types.Schedule, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ScheduleKey)
	bzSchedule := store.Get(types.GetScheduleKey(name))
	if bzSchedule == nil {
		return nil, false
	} else {
		var schedule types.Schedule
		k.cdc.MustUnmarshal(bzSchedule, &schedule)
		return &schedule, true
	}
}

func (k *Keeper) getSchedulesReadyForExecution(ctx sdk.Context) []types.Schedule {
	params := k.GetParams(ctx)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ScheduleKey)
	count := uint64(0)

	res := make([]types.Schedule, 0)

	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var schedule types.Schedule
		k.cdc.MustUnmarshal(iterator.Value(), &schedule)

		if k.intervalPassed(ctx, schedule) {
			res = append(res, schedule)
			count += 1

			if count >= params.Limit {
				return res
			}
		}
	}

	return res
}

func (k *Keeper) executeSchedule(ctx sdk.Context, msgServer wasmtypes.MsgServer, schedule types.Schedule) {
	for _, msg := range schedule.Msgs {
		_, err := msgServer.ExecuteContract(sdk.WrapSDKContext(ctx), &msg)
		if err != nil {
			// TODO: return err, log warn or err?
		}

		// Even if contract execution returned an error, we still increase the height
		// and execute it after this interval
		schedule.LastExecuteHeight = uint64(ctx.BlockHeight()) // TODO: ok conversion?
		k.storeSchedule(ctx, schedule)
	}
}

func (k *Keeper) storeSchedule(ctx sdk.Context, schedule types.Schedule) {
	store := ctx.KVStore(k.storeKey)

	bzSchedule := k.cdc.MustMarshal(&schedule)
	store.Set(types.GetScheduleKey(schedule.Name), bzSchedule)
}

func (k *Keeper) removeSchedule(ctx sdk.Context, name string) {
	store := ctx.KVStore(k.storeKey)

	store.Delete(types.GetScheduleKey(name))
}

func (k *Keeper) intervalPassed(ctx sdk.Context, schedule types.Schedule) bool {
	return uint64(ctx.BlockHeight()) > (schedule.LastExecuteHeight + schedule.Period) // TODO: ok conversion?
}
