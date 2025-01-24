package types_test

import (
	"testing"

	"cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
	"github.com/stretchr/testify/require"
)

func TestDefaultGenesis(t *testing.T) {
	defaultGenesis := revenuetypes.DefaultGenesis()
	require.NotNil(t, defaultGenesis)
	require.Equal(t, revenuetypes.DefaultParams(), defaultGenesis.Params)
	require.Equal(t, 0, len(defaultGenesis.Validators))

	ps, ok := defaultGenesis.State.PaymentSchedule.GetCachedValue().(revenuetypes.PaymentSchedule)
	require.True(t, ok)
	require.Equal(t, &revenuetypes.EmptyPaymentSchedule{}, ps)

	require.Nil(t, defaultGenesis.Validate())
}

func TestInvalidGenesisPaymentScheduleTypeMismatch(t *testing.T) {
	defaultGenesis := revenuetypes.DefaultGenesis()
	defaultGenesis.Params.PaymentScheduleType = revenuetypes.PaymentScheduleType_PAYMENT_SCHEDULE_TYPE_BLOCK_BASED
	err := defaultGenesis.Validate()
	require.ErrorContains(t, err, "does not match payment schedule")

	defaultGenesis = revenuetypes.DefaultGenesis()
	defaultGenesis.State.PaymentSchedule, err = codectypes.NewAnyWithValue(&revenuetypes.MonthlyPaymentSchedule{})
	require.Nil(t, err)
	err = defaultGenesis.Validate()
	require.ErrorContains(t, err, "does not match payment schedule")
}

func TestInvalidGenesisParams(t *testing.T) {
	defaultGenesis := revenuetypes.DefaultGenesis()
	defaultGenesis.Params.PaymentScheduleType = 999
	require.ErrorContains(t, defaultGenesis.Validate(), "invalid payment schedule type")

	defaultGenesis = revenuetypes.DefaultGenesis()
	defaultGenesis.Params.BlocksPerformanceRequirement.AllowedToMiss = math.LegacyOneDec().Add(math.LegacySmallestDec())
	require.ErrorContains(t, defaultGenesis.Validate(), "blocks allowed to miss must be between 0.0 and 1.0")
	defaultGenesis = revenuetypes.DefaultGenesis()
	defaultGenesis.Params.BlocksPerformanceRequirement.AllowedToMiss = math.LegacyZeroDec().Sub(math.LegacySmallestDec())
	require.ErrorContains(t, defaultGenesis.Validate(), "blocks allowed to miss must be between 0.0 and 1.0")

	defaultGenesis = revenuetypes.DefaultGenesis()
	defaultGenesis.Params.BlocksPerformanceRequirement.RequiredAtLeast = math.LegacyOneDec().Add(math.LegacySmallestDec())
	require.ErrorContains(t, defaultGenesis.Validate(), "blocks required at least must be between 0.0 and 1.0")
	defaultGenesis = revenuetypes.DefaultGenesis()
	defaultGenesis.Params.BlocksPerformanceRequirement.RequiredAtLeast = math.LegacyZeroDec().Sub(math.LegacySmallestDec())
	require.ErrorContains(t, defaultGenesis.Validate(), "blocks required at least must be between 0.0 and 1.0")

	defaultGenesis = revenuetypes.DefaultGenesis()
	defaultGenesis.Params.OracleVotesPerformanceRequirement.AllowedToMiss = math.LegacyOneDec().Add(math.LegacySmallestDec())
	require.ErrorContains(t, defaultGenesis.Validate(), "oracle votes allowed to miss must be between 0.0 and 1.0")
	defaultGenesis = revenuetypes.DefaultGenesis()
	defaultGenesis.Params.OracleVotesPerformanceRequirement.AllowedToMiss = math.LegacyZeroDec().Sub(math.LegacySmallestDec())
	require.ErrorContains(t, defaultGenesis.Validate(), "oracle votes allowed to miss must be between 0.0 and 1.0")

	defaultGenesis = revenuetypes.DefaultGenesis()
	defaultGenesis.Params.OracleVotesPerformanceRequirement.RequiredAtLeast = math.LegacyOneDec().Add(math.LegacySmallestDec())
	require.ErrorContains(t, defaultGenesis.Validate(), "oracle votes required at least must be between 0.0 and 1.0")
	defaultGenesis = revenuetypes.DefaultGenesis()
	defaultGenesis.Params.OracleVotesPerformanceRequirement.RequiredAtLeast = math.LegacyZeroDec().Sub(math.LegacySmallestDec())
	require.ErrorContains(t, defaultGenesis.Validate(), "oracle votes required at least must be between 0.0 and 1.0")
}
