package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	appconfig "github.com/neutron-org/neutron/v6/app/config"
	"github.com/neutron-org/neutron/v6/app/params"
	mock_types "github.com/neutron-org/neutron/v6/testutil/mocks/revenue/types"
	testutil_keeper "github.com/neutron-org/neutron/v6/testutil/revenue/keeper"
	revenuekeeper "github.com/neutron-org/neutron/v6/x/revenue/keeper"
	revenuetypes "github.com/neutron-org/neutron/v6/x/revenue/types"
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
			"ValidParams",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					TwapWindow:  3600 * 24,
					RewardAsset: params.DefaultDenom,
					RewardQuote: &revenuetypes.RewardQuote{Asset: "USD", Amount: 1500},
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
			"EmptyAuthority",
			&revenuetypes.MsgUpdateParams{
				Authority: "",
				Params: revenuetypes.Params{
					TwapWindow:  3600 * 24,
					RewardAsset: params.DefaultDenom,
					RewardQuote: &revenuetypes.RewardQuote{Asset: "USD", Amount: 1500},
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
			"Unauthorized",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron1hxskfdxpp5hqgtjj6am6nkjefhfzj359x0ar3z",
				Params: revenuetypes.Params{
					TwapWindow:  3600 * 24,
					RewardAsset: params.DefaultDenom,
					RewardQuote: &revenuetypes.RewardQuote{Asset: "USD", Amount: 1500},
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
			"TooLowPerformanceRequirement",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					TwapWindow:  3600 * 24,
					RewardAsset: params.DefaultDenom,
					RewardQuote: &revenuetypes.RewardQuote{Asset: "USD", Amount: 1500},
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
			"TooBigPerformanceRequirement",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					TwapWindow:  3600 * 24,
					RewardAsset: params.DefaultDenom,
					RewardQuote: &revenuetypes.RewardQuote{Asset: "USD", Amount: 1500},
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
			"InvalidSumOfBlockPerformanceRequirementParams",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					TwapWindow:  3600 * 24,
					RewardAsset: params.DefaultDenom,
					RewardQuote: &revenuetypes.RewardQuote{Asset: "USD", Amount: 1500},
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
			"InvalidSumOfOracleVotesPerformanceRequirementParams",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					TwapWindow:  3600 * 24,
					RewardAsset: params.DefaultDenom,
					RewardQuote: &revenuetypes.RewardQuote{Asset: "USD", Amount: 1500},
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
			"InvalidSumOfOracleVotesPerformanceRequirementParams",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					TwapWindow:  3600 * 24,
					RewardAsset: params.DefaultDenom,
					RewardQuote: &revenuetypes.RewardQuote{Asset: "USD", Amount: 1500},
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
			"UnitilializedEmptyPaymentScheduleType",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					TwapWindow:  3600 * 24,
					RewardAsset: params.DefaultDenom,
					RewardQuote: &revenuetypes.RewardQuote{Asset: "USD", Amount: 1500},
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
			"UnitilializedMonthlyPaymentScheduleType",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					TwapWindow:  3600 * 24,
					RewardAsset: params.DefaultDenom,
					RewardQuote: &revenuetypes.RewardQuote{Asset: "USD", Amount: 1500},
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
			"UnitilializedBlockBasedPaymentScheduleTypeNilInnerValue",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					TwapWindow:  3600 * 24,
					RewardAsset: params.DefaultDenom,
					RewardQuote: &revenuetypes.RewardQuote{Asset: "USD", Amount: 1500},
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
			"UnitilializedBlockBasedPaymentScheduleTypeZeroBlocksPerPeriod",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					TwapWindow:  3600 * 24,
					RewardAsset: params.DefaultDenom,
					RewardQuote: &revenuetypes.RewardQuote{Asset: "USD", Amount: 1500},
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
		{
			"InvalidTwapWindowTooLow",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					TwapWindow:  0,
					RewardAsset: params.DefaultDenom,
					RewardQuote: &revenuetypes.RewardQuote{Asset: "USD", Amount: 1500},
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
			"params are invalid: twap window must be between 1 and 2592000",
		},
		{
			"InvalidTwapWindowTooHigh",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					TwapWindow:  revenuetypes.MaxTWAPWindow + 1,
					RewardAsset: params.DefaultDenom,
					RewardQuote: &revenuetypes.RewardQuote{Asset: "USD", Amount: 1500},
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
			"params are invalid: twap window must be between 1 and 2592000",
		},
		{
			"ProhibitedRewardAssetChange",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					TwapWindow:  3600 * 24,
					RewardAsset: "uatom",
					RewardQuote: &revenuetypes.RewardQuote{Asset: "USD", Amount: 1500},
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
			"reward asset change is prohibited",
		},
		{
			"ProhibitedQuoteAssetChange",
			&revenuetypes.MsgUpdateParams{
				Authority: "neutron159kr6k0y4f43dsrdyqlm9x23jajunegal4nglw044u7zl72u0eeqharq3a",
				Params: revenuetypes.Params{
					TwapWindow:  3600 * 24,
					RewardAsset: params.DefaultDenom,
					RewardQuote: &revenuetypes.RewardQuote{Asset: "BTC", Amount: 1500},
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
			"quote asset change is prohibited",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := msgServer.UpdateParams(ctx, tt.updateParamsMsg)

			if tt.expectedErr == "" {
				require.NoError(t, err, tt.expectedErr)
				require.Equal(t, res, &revenuetypes.MsgUpdateParamsResponse{})
			} else {
				require.ErrorContains(t, err, tt.expectedErr)
				require.Empty(t, res)
			}
		})
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
