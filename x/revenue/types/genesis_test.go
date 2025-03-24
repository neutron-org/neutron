package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	revenuetypes "github.com/neutron-org/neutron/v6/x/revenue/types"
)

func TestDefaultGenesis(t *testing.T) {
	defaultGenesis := revenuetypes.DefaultGenesis()
	require.NotNil(t, defaultGenesis)
	require.Equal(t, revenuetypes.DefaultParams(), defaultGenesis.Params)
	require.Equal(t, 0, len(defaultGenesis.Validators))

	psi, err := defaultGenesis.PaymentSchedule.IntoPaymentScheduleI()
	require.Nil(t, err)
	require.Equal(t, &revenuetypes.EmptyPaymentSchedule{}, psi)

	require.Nil(t, defaultGenesis.Validate())
}

func TestInvalidGenesisPaymentScheduleTypeMismatch(t *testing.T) {
	defaultGenesis := revenuetypes.DefaultGenesis()
	defaultGenesis.Params.PaymentScheduleType = &revenuetypes.PaymentScheduleType{
		PaymentScheduleType: &revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{
			BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 1},
		},
	}
	err := defaultGenesis.Validate()
	require.ErrorContains(t, err, "does not match payment schedule")

	defaultGenesis = revenuetypes.DefaultGenesis()
	defaultGenesis.PaymentSchedule = (&revenuetypes.MonthlyPaymentSchedule{}).IntoPaymentSchedule()
	err = defaultGenesis.Validate()
	require.ErrorContains(t, err, "does not match payment schedule")
}

func TestInvalidGenesisParams(t *testing.T) {
	defaultGenesis := revenuetypes.DefaultGenesis()
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

	defaultGenesis = revenuetypes.DefaultGenesis()
	defaultGenesis.Params.BlocksPerformanceRequirement.AllowedToMiss = math.LegacyOneDec()
	defaultGenesis.Params.BlocksPerformanceRequirement.RequiredAtLeast = math.LegacySmallestDec()
	require.ErrorContains(t, defaultGenesis.Validate(), "sum of blocks allowed to miss and required at least must not be greater than 1.0")
	defaultGenesis = revenuetypes.DefaultGenesis()
	defaultGenesis.Params.OracleVotesPerformanceRequirement.AllowedToMiss = math.LegacyOneDec()
	defaultGenesis.Params.OracleVotesPerformanceRequirement.RequiredAtLeast = math.LegacySmallestDec()
	require.ErrorContains(t, defaultGenesis.Validate(), "sum of oracle votes allowed to miss and required at least must not be greater than 1.0")

	defaultGenesis = revenuetypes.DefaultGenesis()
	defaultGenesis.Params.PaymentScheduleType = &revenuetypes.PaymentScheduleType{
		PaymentScheduleType: &revenuetypes.PaymentScheduleType_MonthlyPaymentScheduleType{
			MonthlyPaymentScheduleType: &revenuetypes.MonthlyPaymentScheduleType{},
		},
	}
	require.ErrorContains(t, defaultGenesis.Validate(), "payment schedule type *types.PaymentScheduleType_MonthlyPaymentScheduleType does not match payment schedule of type *types.EmptyPaymentSchedule in genesis state")

	defaultGenesis = revenuetypes.DefaultGenesis()
	defaultGenesis.PaymentSchedule = (&revenuetypes.MonthlyPaymentSchedule{}).IntoPaymentSchedule()
	require.ErrorContains(t, defaultGenesis.Validate(), "payment schedule type *types.PaymentScheduleType_EmptyPaymentScheduleType does not match payment schedule of type *types.MonthlyPaymentSchedule in genesis state")
}
