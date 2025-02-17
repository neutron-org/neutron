package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	appconfig "github.com/neutron-org/neutron/v5/app/config"
	mock_types "github.com/neutron-org/neutron/v5/testutil/mocks/revenue/types"
	testutil_keeper "github.com/neutron-org/neutron/v5/testutil/revenue/keeper"
	revenuekeeper "github.com/neutron-org/neutron/v5/x/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
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
			"valid params",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					BaseCompensation: 1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.PaymentScheduleType{
						PaymentScheduleType: &revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{
							BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 10},
						},
					},
				},
			},
			"",
		},
		{
			"empty authority",
			&revenuetypes.MsgUpdateParams{
				Authority: "",
				Params: revenuetypes.Params{
					BaseCompensation: 1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.PaymentScheduleType{
						PaymentScheduleType: &revenuetypes.PaymentScheduleType_EmptyPaymentScheduleType{
							EmptyPaymentScheduleType: &revenuetypes.EmptyPaymentScheduleType{},
						},
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
					BaseCompensation: 1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.PaymentScheduleType{
						PaymentScheduleType: &revenuetypes.PaymentScheduleType_EmptyPaymentScheduleType{
							EmptyPaymentScheduleType: &revenuetypes.EmptyPaymentScheduleType{},
						},
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
					BaseCompensation: 1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec().Sub(math.LegacyOneDec()),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.PaymentScheduleType{
						PaymentScheduleType: &revenuetypes.PaymentScheduleType_EmptyPaymentScheduleType{
							EmptyPaymentScheduleType: &revenuetypes.EmptyPaymentScheduleType{},
						},
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
					BaseCompensation: 1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec().Add(math.LegacyOneDec()),
					},
					PaymentScheduleType: &revenuetypes.PaymentScheduleType{
						PaymentScheduleType: &revenuetypes.PaymentScheduleType_EmptyPaymentScheduleType{
							EmptyPaymentScheduleType: &revenuetypes.EmptyPaymentScheduleType{},
						},
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
					BaseCompensation: 1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyOneDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.PaymentScheduleType{
						PaymentScheduleType: &revenuetypes.PaymentScheduleType_EmptyPaymentScheduleType{
							EmptyPaymentScheduleType: &revenuetypes.EmptyPaymentScheduleType{},
						},
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
					BaseCompensation: 1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyOneDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.PaymentScheduleType{
						PaymentScheduleType: &revenuetypes.PaymentScheduleType_EmptyPaymentScheduleType{
							EmptyPaymentScheduleType: &revenuetypes.EmptyPaymentScheduleType{},
						},
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
					BaseCompensation: 1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyOneDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.PaymentScheduleType{
						PaymentScheduleType: &revenuetypes.PaymentScheduleType_EmptyPaymentScheduleType{
							EmptyPaymentScheduleType: &revenuetypes.EmptyPaymentScheduleType{},
						},
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
					BaseCompensation: 1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.PaymentScheduleType{
						PaymentScheduleType: &revenuetypes.PaymentScheduleType_EmptyPaymentScheduleType{},
					},
				},
			},
			"inner empty payment schedule is nil",
		},
		{
			"unitilialized monthly payment schedule type",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					BaseCompensation: 1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.PaymentScheduleType{
						PaymentScheduleType: &revenuetypes.PaymentScheduleType_MonthlyPaymentScheduleType{},
					},
				},
			},
			"inner monthly payment schedule is nil",
		},
		{
			"unitilialized block based payment schedule type: nil inner value",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					BaseCompensation: 1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.PaymentScheduleType{
						PaymentScheduleType: &revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{},
					},
				},
			},
			"inner block based payment schedule is nil",
		},
		{
			"unitilialized block based payment schedule type : zero blocks per period",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					BaseCompensation: 1500,
					BlocksPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					OracleVotesPerformanceRequirement: &revenuetypes.PerformanceRequirement{
						AllowedToMiss:   math.LegacyZeroDec(),
						RequiredAtLeast: math.LegacyOneDec(),
					},
					PaymentScheduleType: &revenuetypes.PaymentScheduleType{
						PaymentScheduleType: &revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{
							BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 0},
						},
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

func TestFundTreasury(t *testing.T) {
	appconfig.GetDefaultConfig()

	ctrl := gomock.NewController(t)
	bankKeeper := mock_types.NewMockBankKeeper(ctrl)

	k, ctx := testutil_keeper.RevenueKeeper(t, bankKeeper, nil, "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a")
	require.Nil(t, k.SetParams(ctx, revenuetypes.DefaultParams()))
	msgServer := revenuekeeper.NewMsgServerImpl(k)

	tests := []struct {
		name            string
		fundTreasuryMsg *revenuetypes.MsgFundTreasury
		expectedErr     string
	}{
		{
			"valid top up",
			&revenuetypes.MsgFundTreasury{
				Sender: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Amount: sdktypes.NewCoins(sdktypes.NewCoin("untrn", math.NewInt(1000))),
			},
			"",
		},
		{
			"too many coins",
			&revenuetypes.MsgFundTreasury{
				Sender: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Amount: sdktypes.NewCoins(
					sdktypes.NewCoin("untrn", math.NewInt(1000)),
					sdktypes.NewCoin("uatom", math.NewInt(1000)),
				),
			},
			"exactly one coin must be provided",
		},
		{
			"invalid denom",
			&revenuetypes.MsgFundTreasury{
				Sender: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Amount: sdktypes.NewCoins(sdktypes.NewCoin("uatom", math.NewInt(1000))),
			},
			"provided denom doesn't match the reward denom untrn",
		},
	}

	for _, tt := range tests {
		if tt.expectedErr == "" {
			bankKeeper.EXPECT().SendCoinsFromAccountToModule(
				gomock.Any(),
				sdktypes.MustAccAddressFromBech32(tt.fundTreasuryMsg.Sender),
				revenuetypes.RevenueTreasuryPoolName,
				tt.fundTreasuryMsg.Amount,
			).Times(1) // expect a single transfer from sender to module account

			res, err := msgServer.FundTreasury(ctx, tt.fundTreasuryMsg)
			require.NoError(t, err, tt.expectedErr)
			require.Equal(t, res, &revenuetypes.MsgFundTreasuryResponse{})
		} else {
			res, err := msgServer.FundTreasury(ctx, tt.fundTreasuryMsg)
			require.ErrorContains(t, err, tt.expectedErr)
			require.Empty(t, res)
		}
	}
}
