package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/neutron-org/neutron/x/incentives/types"
	"github.com/stretchr/testify/assert"
)

func TestGaugeIsUpcomingGauge(t *testing.T) {
	now := time.Now()
	gauge := NewGauge(1, false, QueryCondition{}, sdk.Coins{}, now.Add(time.Minute), 10, 0, sdk.Coins{}, 0)

	assert.True(t, gauge.IsUpcomingGauge(now))
	assert.False(t, gauge.IsUpcomingGauge(now.Add(time.Minute)))
}

func TestGaugeIsActiveGauge(t *testing.T) {
	now := time.Now()
	gauge := NewGauge(1, false, QueryCondition{}, sdk.Coins{}, now.Add(time.Minute), 10, 0, sdk.Coins{}, 0)

	assert.False(t, gauge.IsActiveGauge(now))
	assert.True(t, gauge.IsActiveGauge(now.Add(11*time.Minute)))

	gauge.IsPerpetual = true
	assert.False(t, gauge.IsActiveGauge(now))
	assert.True(t, gauge.IsActiveGauge(now.Add(11*time.Minute)))
}

func TestGaugeIsFinishedGauge(t *testing.T) {
	now := time.Now()
	gauge := NewGauge(1, false, QueryCondition{}, sdk.Coins{}, now.Add(-time.Minute), 10, 10, sdk.Coins{}, 0)
	assert.True(t, gauge.IsFinishedGauge(now))

	gauge = NewGauge(1, false, QueryCondition{}, sdk.Coins{}, now.Add(-time.Minute), 10, 8, sdk.Coins{}, 0)
	assert.False(t, gauge.IsFinishedGauge(now))
}

func TestGaugeEpochsRemaining(t *testing.T) {
	gauge := NewGauge(1, false, QueryCondition{}, sdk.Coins{}, time.Time{}, 10, 5, sdk.Coins{}, 0)

	assert.Equal(t, uint64(5), gauge.EpochsRemaining())

	gauge.IsPerpetual = true
	assert.Equal(t, uint64(1), gauge.EpochsRemaining())
}

func TestGaugeCoinsRemaining(t *testing.T) {
	coins := sdk.Coins{sdk.NewCoin("coin1", sdk.NewInt(100))}
	distCoins := sdk.Coins{sdk.NewCoin("coin1", sdk.NewInt(50))}
	gauge := NewGauge(1, false, QueryCondition{}, coins, time.Time{}, 10, 5, distCoins, 0)
	assert.Equal(t, sdk.Coins{sdk.NewCoin("coin1", sdk.NewInt(50))}, gauge.CoinsRemaining())
}

func TestGaugeGetTotal(t *testing.T) {
	distSpec := DistributionSpec{
		"addr1": sdk.Coins{sdk.NewCoin("coin1", sdk.NewInt(10))},
		"addr2": sdk.Coins{sdk.NewCoin("coin1", sdk.NewInt(20))},
		"addr3": sdk.Coins{sdk.NewCoin("coin1", sdk.NewInt(30))},
	}

	assert.Equal(t, sdk.Coins{sdk.NewCoin("coin1", sdk.NewInt(60))}, distSpec.GetTotal())
}
