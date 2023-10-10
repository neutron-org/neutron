package types

import (
	time "time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGauge creates a new gauge struct given the required gauge parameters.
func NewGauge(
	id uint64,
	isPerpetual bool,
	distrTo QueryCondition,
	coins sdk.Coins,
	startTime time.Time,
	numEpochsPaidOver uint64,
	filledEpochs uint64,
	distCoins sdk.Coins,
	pricingTick int64,
) Gauge {
	return Gauge{
		Id:                id,
		IsPerpetual:       isPerpetual,
		DistributeTo:      distrTo,
		Coins:             coins,
		StartTime:         startTime,
		NumEpochsPaidOver: numEpochsPaidOver,
		FilledEpochs:      filledEpochs,
		DistributedCoins:  distCoins,
		PricingTick:       pricingTick,
	}
}

func (gauge Gauge) hasEpochsRemaining() bool {
	return gauge.IsPerpetual || gauge.FilledEpochs < gauge.NumEpochsPaidOver
}

func (gauge Gauge) hasStarted(now time.Time) bool {
	return !now.Before(gauge.StartTime)
}

// IsUpcomingGauge returns true if the gauge's distribution start time is after the provided time.
func (gauge Gauge) IsUpcomingGauge(now time.Time) bool {
	return !gauge.hasStarted(now) && gauge.hasEpochsRemaining()
}

// IsActiveGauge returns true if the gauge is in an active state during the provided time.
func (gauge Gauge) IsActiveGauge(now time.Time) bool {
	return gauge.hasStarted(now) && gauge.hasEpochsRemaining()
}

// IsFinishedGauge returns true if the gauge is in a finished state during the provided time.
func (gauge Gauge) IsFinishedGauge(now time.Time) bool {
	return gauge.hasStarted(now) && !gauge.hasEpochsRemaining()
}

func (gauge Gauge) RewardsNextEpoch() sdk.Coins {
	result := sdk.Coins{}
	epochsRemaining := gauge.EpochsRemaining()
	if epochsRemaining == 0 {
		return result
	}

	for _, rewardRemainingCoin := range gauge.CoinsRemaining() {
		amount := rewardRemainingCoin.Amount.Quo(math.NewInt(int64(epochsRemaining)))
		result = result.Add(sdk.Coin{Denom: rewardRemainingCoin.Denom, Amount: amount})
	}

	return result
}

func (gauge Gauge) EpochsRemaining() uint64 {
	if !gauge.IsPerpetual {
		return gauge.NumEpochsPaidOver - gauge.FilledEpochs
	}

	return 1
}

func (gauge Gauge) CoinsRemaining() sdk.Coins {
	return gauge.Coins.Sub(gauge.DistributedCoins...)
}
