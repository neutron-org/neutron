package keeper

import (
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm"
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
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
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
	}
}

// TODO: DOC comments

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k *Keeper) CheckTimer(ctx sdk.Context) {
	// TODO
}

// period in blocks
func (k *Keeper) AddSchedule(ctx sdk.Context, name string, period uint64, msgs []wasm.MsgExecuteContract) {

}

func (k *Keeper) RemoveSchedule(ctx sdk.Context, name string) {

}

func (k *Keeper) ExecuteSchedule() {
	// TODO
}

func (k Keeper) GetAllSchedules(ctx sdk.Context) []types.Schedule {
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

func (k Keeper) storeSchedule(ctx sdk.Context, schedule types.Schedule) {
	store := ctx.KVStore(k.storeKey)

	bzSchedule := k.cdc.MustMarshal(&schedule)
	store.Set(types.GetScheduleKey(schedule.name), bzSchedule)
}

func (k Keeper) removeSchedule(ctx sdk.Context, name string) {
	store := ctx.KVStore(k.storeKey)

	store.Delete(types.GetScheduleKey(name))
}
