package keeper

import (
	"fmt"
	"math/big"

	"github.com/armon/go-metrics"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/telemetry"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v3/x/dex/types"
)

var (
	MetricLabelWithdrawn          = "total_withdrawn"
	MetricLabelGasConsumed        = "gas_consumed"
	MetricLabelTotalOrdersExpired = "total_orders_expired"
	MetricLabelTotalLimitOrders   = "total_orders_limit"
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

func incWithdrawnAmount(ctx sdk.Context, coins sdk.Coins) {
	for _, coin := range coins {
		divisor := big.NewInt(6)
		f, _ := new(big.Int).Div(coin.Amount.BigInt(), divisor).Float64() // todo check err
		telemetry.IncrCounterWithLabels([]string{MetricLabelWithdrawn}, float32(f), []metrics.Label{
			telemetry.NewLabel(telemetry.MetricLabelNameModule, types.ModuleName),
			telemetry.NewLabel("denom", coin.Denom),
		})
	}
}

func gasConsumed(ctx sdk.Context) {
	gas := ctx.GasMeter().GasConsumed()
	gasFloat := float32(gas)
	telemetry.SetGaugeWithLabels([]string{MetricLabelWithdrawn}, gasFloat, []metrics.Label{
		telemetry.NewLabel(telemetry.MetricLabelNameModule, types.ModuleName),
		telemetry.NewLabel(MetricLabelGasConsumed, "todo"),
	})
}

func incExpiredOdrers() {
	telemetry.IncrCounterWithLabels([]string{MetricLabelTotalOrdersExpired}, float32(1), []metrics.Label{
		telemetry.NewLabel(telemetry.MetricLabelNameModule, types.ModuleName),
	})
}

func totalLimitOrders() {
	telemetry.IncrCounterWithLabels([]string{MetricLabelTotalOrdersExpired}, float32(1), []metrics.Label{
		telemetry.NewLabel(telemetry.MetricLabelNameModule, types.ModuleName),
	})
}
