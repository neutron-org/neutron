package keeper

import (
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v4/x/dynamicfees/types"
)

var (
	LabelExecuteReadySchedules   = "execute_ready_schedules"
	LabelScheduleCount           = "schedule_count"
	LabelScheduleExecutionsCount = "schedule_executions_count"

	MetricLabelSuccess      = "success"
	MetricLabelScheduleName = "schedule_name"
)

type (
	Keeper struct {
		cdc       codec.BinaryCodec
		storeKey  storetypes.StoreKey
		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		authority: authority,
	}
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
