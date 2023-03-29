package keeper

import (
	"fmt"
	"strconv"
	"time"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/neutron-org/neutron/x/cron/types"
	"github.com/tendermint/tendermint/libs/log"
)

var (
	LabelCheckTimer    = "check_timer"
	MetricLabelSuccess = "success"
	MetricMsgIndex     = "msg_idx"
	MetricScheduleName = "schedule_name"
)

type (
	Keeper struct {
		cdc           codec.BinaryCodec
		storeKey      storetypes.StoreKey
		memKey        storetypes.StoreKey
		paramstore    paramtypes.Subspace
		accountKeeper types.AccountKeeper
		WasmMsgServer wasmtypes.MsgServer
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		paramstore:    ps,
		accountKeeper: accountKeeper,
	}
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ExecuteReadySchedules gets all schedules that are due for execution (with limit that is equals to Params.Limit)
// and executes messages in each one
// NOTE that errors in contract calls do NOT stop schedule execution
func (k *Keeper) ExecuteReadySchedules(ctx sdk.Context) {
	telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), LabelCheckTimer)
	schedules := k.getSchedulesReadyForExecution(ctx)

	for _, schedule := range schedules {
		k.executeSchedule(ctx, schedule)
	}
}

// AddSchedule adds new schedule to execution for every block `period`.
// First schedule execution is supposed to be on `now + period` block.
func (k *Keeper) AddSchedule(ctx sdk.Context, name string, period uint64, msgs []types.MsgExecuteContract) {
	schedule := types.Schedule{
		Name:              name,
		Period:            period,
		Msgs:              msgs,
		LastExecuteHeight: uint64(ctx.BlockHeight()), // let's execute newly added schedule on `now + period` block
	}
	k.storeSchedule(ctx, schedule)
}

// RemoveSchedule removes schedule with a given `name`
func (k *Keeper) RemoveSchedule(ctx sdk.Context, name string) {
	k.removeSchedule(ctx, name)
}

// GetAllSchedules returns all schedules
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

// GetSchedule returns schedule with a given `name`
func (k *Keeper) GetSchedule(ctx sdk.Context, name string) (*types.Schedule, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ScheduleKey)
	bzSchedule := store.Get(types.GetScheduleKey(name))
	if bzSchedule == nil {
		return nil, false
	}

	var schedule types.Schedule
	k.cdc.MustUnmarshal(bzSchedule, &schedule)
	return &schedule, true
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
			count++

			if count >= params.Limit {
				return res
			}
		}
	}

	return res
}

// executeSchedule executes given schedule and changes LastExecuteHeight
func (k *Keeper) executeSchedule(ctx sdk.Context, schedule types.Schedule) {
	for idx, msg := range schedule.Msgs {
		executeMsg := wasmtypes.MsgExecuteContract{
			Sender:   k.accountKeeper.GetModuleAddress(types.ModuleName).String(), // TODO: store in constructor to avoid calculating every time?
			Contract: msg.Contract,
			Msg:      msg.Msg,
			Funds:    sdk.NewCoins(),
		}
		_, err := k.WasmMsgServer.ExecuteContract(sdk.WrapSDKContext(ctx), &executeMsg)

		countMsgExecuted(err, schedule.Name, idx)

		if err != nil {
			ctx.Logger().Info("executeSchedule: failed to execute contract msg",
				"schedule_name", schedule.Name,
				"msg_idx", idx,
				"msg_contract", msg.Contract,
				"error", err,
			)
		}

		// Even if contract execution returned an error, we still increase the height
		// and execute it after this interval
		schedule.LastExecuteHeight = uint64(ctx.BlockHeight())
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
	return uint64(ctx.BlockHeight()) > (schedule.LastExecuteHeight + schedule.Period)
}

func countMsgExecuted(err error, scheduleName string, idx int) {
	telemetry.IncrCounterWithLabels([]string{"execute_schedule"}, 1, []metrics.Label{
		telemetry.NewLabel(telemetry.MetricLabelNameModule, types.ModuleName),
		telemetry.NewLabel(MetricScheduleName, scheduleName),
		telemetry.NewLabel(MetricLabelSuccess, strconv.FormatBool(err == nil)),
		telemetry.NewLabel(MetricMsgIndex, strconv.Itoa(idx)),
	})
}
