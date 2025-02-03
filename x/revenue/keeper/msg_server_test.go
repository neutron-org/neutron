package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	appconfig "github.com/neutron-org/neutron/v5/app/config"
	testutil_keeper "github.com/neutron-org/neutron/v5/testutil/revenue/keeper"
	revenuekeeper "github.com/neutron-org/neutron/v5/x/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
	"github.com/stretchr/testify/require"
)

func TestUpdateParams(t *testing.T) {
	appconfig.GetDefaultConfig()
	k, ctx := testutil_keeper.RevenueKeeper(t, nil, nil, "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a")
	msgServer := revenuekeeper.NewMsgServerImpl(k)

	tests := []struct {
		name            string
		updateParamsMsg *revenuetypes.MsgUpdateParams
		expectedErr     string
	}{
		{
			"empty authority",
			&revenuetypes.MsgUpdateParams{
				Authority: "",
				Params: revenuetypes.Params{
					DenomCompensation: "untrn",
					BaseCompensation:  1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.Params_EmptyPaymentScheduleType{
						EmptyPaymentScheduleType: &revenuetypes.EmptyPaymentScheduleType{},
					},
				},
			},
			"authority is invalid: empty address string is not allowed",
		},
		{
			"unauthorized",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z",
				Params: revenuetypes.Params{
					DenomCompensation: "untrn",
					BaseCompensation:  1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.Params_EmptyPaymentScheduleType{
						EmptyPaymentScheduleType: &revenuetypes.EmptyPaymentScheduleType{},
					},
				},
			},
			"invalid authority; expected neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a, got neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z",
		},
		{
			"too low performance requirement",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					DenomCompensation: "untrn",
					BaseCompensation:  1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec().Sub(math.LegacyOneDec()),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.Params_EmptyPaymentScheduleType{
						EmptyPaymentScheduleType: &revenuetypes.EmptyPaymentScheduleType{},
					},
				},
			},
			"blocks allowed to miss must be between 0.0 and 1.0",
		},
		{
			"too big performance requirement",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					DenomCompensation: "untrn",
					BaseCompensation:  1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec().Add(math.LegacyOneDec()),
					},
					PaymentScheduleType: &revenuetypes.Params_EmptyPaymentScheduleType{
						EmptyPaymentScheduleType: &revenuetypes.EmptyPaymentScheduleType{},
					},
				},
			},
			"oracle votes required at least must be between 0.0 and 1.0",
		},
		{
			"invalid sum of block performance requirement params",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					DenomCompensation: "untrn",
					BaseCompensation:  1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyOneDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.Params_EmptyPaymentScheduleType{
						EmptyPaymentScheduleType: &revenuetypes.EmptyPaymentScheduleType{},
					},
				},
			},
			"sum of blocks allowed to miss and required at least must not be greater than 1.0",
		},
		{
			"invalid sum of oracle votes performance requirement params",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					DenomCompensation: "untrn",
					BaseCompensation:  1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyOneDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.Params_EmptyPaymentScheduleType{
						EmptyPaymentScheduleType: &revenuetypes.EmptyPaymentScheduleType{},
					},
				},
			},
			"sum of oracle votes allowed to miss and required at least must not be greater than 1.0",
		},
		{
			"invalid sum of oracle votes performance requirement params",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					DenomCompensation: "untrn",
					BaseCompensation:  1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyOneDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.Params_EmptyPaymentScheduleType{
						EmptyPaymentScheduleType: &revenuetypes.EmptyPaymentScheduleType{},
					},
				},
			},
			"sum of oracle votes allowed to miss and required at least must not be greater than 1.0",
		},
		{
			"unitilialized empty payment schedule type",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					DenomCompensation: "untrn",
					BaseCompensation:  1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.Params_EmptyPaymentScheduleType{},
				},
			},
			"inner empty payment schedule is nil",
		},
		{
			"unitilialized monthly payment schedule type",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					DenomCompensation: "untrn",
					BaseCompensation:  1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.Params_MonthlyPaymentScheduleType{},
				},
			},
			"inner monthly payment schedule is nil",
		},
		{
			"unitilialized block based payment schedule type: nil inner value",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					DenomCompensation: "untrn",
					BaseCompensation:  1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.Params_BlockBasedPaymentScheduleType{},
				},
			},
			"inner block based payment schedule is nil",
		},
		{
			"unitilialized block based payment schedule type : zero blocks per period",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					DenomCompensation: "untrn",
					BaseCompensation:  1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.Params_BlockBasedPaymentScheduleType{
						BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 0},
					},
				},
			},
			"block based payment schedule type has zero blocks per period",
		},
	}

	for _, tt := range tests {
		res, err := msgServer.UpdateParams(ctx, tt.updateParamsMsg)

		if tt.expectedErr == "" {
			require.NoError(t, err, tt.expectedErr)
			require.Equal(t, res, &revenuetypes.MsgUpdateParamsResponse{})
		} else {
			require.ErrorContains(t, err, tt.expectedErr)
			require.Empty(t, res)
		}
	}
}
