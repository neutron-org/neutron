package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	revenuetypes "github.com/neutron-org/neutron/v6/x/revenue/types"

	mock_types "github.com/neutron-org/neutron/v6/testutil/mocks/feeburner/types"
	"github.com/neutron-org/neutron/v6/x/feeburner/keeper"

	feekeeperutil "github.com/neutron-org/neutron/v6/testutil/feeburner/keeper"
	feetypes "github.com/neutron-org/neutron/v6/x/feeburner/types"
)

func TestKeeper_RecordBurnedFees(t *testing.T) {
	for _, tc := range []struct {
		desc      string
		initial   *sdk.Coin
		amount    sdk.Coin
		expected  *sdk.Coin
		wantPanic bool
	}{
		{
			desc:      "default works with empty store",
			initial:   nil,
			amount:    sdk.Coin{Denom: feetypes.DefaultNeutronDenom, Amount: math.NewInt(300)},
			expected:  &sdk.Coin{Denom: feetypes.DefaultNeutronDenom, Amount: math.NewInt(300)},
			wantPanic: false,
		},
		{
			desc:      "default works with existing tokens burned sums",
			initial:   &sdk.Coin{Denom: feetypes.DefaultNeutronDenom, Amount: math.NewInt(100)},
			amount:    sdk.Coin{Denom: feetypes.DefaultNeutronDenom, Amount: math.NewInt(300)},
			expected:  &sdk.Coin{Denom: feetypes.DefaultNeutronDenom, Amount: math.NewInt(400)},
			wantPanic: false,
		},
		{
			desc:      "with non-default denom should not write it to the store",
			initial:   nil,
			amount:    sdk.Coin{Denom: "nondefaultdenom", Amount: math.NewInt(300)},
			expected:  nil,
			wantPanic: true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tc.wantPanic {
					t.Errorf("recover = %v, wantPanic = %v", r, tc.wantPanic)
				}
			}()

			feeKeeper, ctx := feekeeperutil.FeeburnerKeeper(t)
			if tc.initial != nil {
				feeKeeper.RecordBurnedFees(ctx, *tc.initial)
			}
			feeKeeper.RecordBurnedFees(ctx, tc.amount)
			res := feeKeeper.GetTotalBurnedNeutronsAmount(ctx)
			require.Equal(t, *tc.expected, res.Coin)
		})
	}
}

func TestKeeper_GetTotalBurnedNeutronsAmount(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		initial  *sdk.Coin
		expected *sdk.Coin
	}{
		{
			desc:     "works with empty value",
			initial:  nil,
			expected: &sdk.Coin{Denom: feetypes.DefaultNeutronDenom, Amount: math.NewInt(0)},
		},
		{
			desc:     "works with existing burned fees",
			initial:  &sdk.Coin{Denom: feetypes.DefaultNeutronDenom, Amount: math.NewInt(100)},
			expected: &sdk.Coin{Denom: feetypes.DefaultNeutronDenom, Amount: math.NewInt(100)},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			feeKeeper, ctx := feekeeperutil.FeeburnerKeeper(t)
			if tc.initial != nil {
				feeKeeper.RecordBurnedFees(ctx, *tc.initial)
			}
			res := feeKeeper.GetTotalBurnedNeutronsAmount(ctx)
			require.Equal(t, tc.expected, &res.Coin)
		})
	}
}

func TestKeeper_BurnAndDistribute_Clean(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	feeKeeper, ctx, _, _ := setupBurnAndDistribute(t, ctrl, sdk.Coins{})

	feeKeeper.BurnAndDistribute(ctx)

	burnedAmount := feeKeeper.GetTotalBurnedNeutronsAmount(ctx)
	require.Equal(t, burnedAmount.Coin.Amount, math.NewInt(0))
}

func TestKeeper_BurnAndDistribute_Ntrn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	feeKeeper, ctx, mockBankKeeper, _ := setupBurnAndDistribute(t, ctrl, sdk.Coins{sdk.NewCoin(feetypes.DefaultNeutronDenom, math.NewInt(100))})

	mockBankKeeper.EXPECT().BurnCoins(ctx, revenuetypes.RevenueFeeRedistributePoolName, sdk.Coins{sdk.NewCoin(feetypes.DefaultNeutronDenom, math.NewInt(100))})

	feeKeeper.BurnAndDistribute(ctx)

	burnedAmount := feeKeeper.GetTotalBurnedNeutronsAmount(ctx)
	require.Equal(t, burnedAmount.Coin.Amount, math.NewInt(100))
}

func TestKeeper_BurnAndDistribute_NonNtrn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	feeKeeper, ctx, mockBankKeeper, redistrAddr := setupBurnAndDistribute(t, ctrl, sdk.Coins{sdk.NewCoin("nonntrn", math.NewInt(50))})

	mockBankKeeper.EXPECT().SendCoins(ctx, redistrAddr, sdk.MustAccAddressFromBech32(feeKeeper.GetParams(ctx).TreasuryAddress), sdk.Coins{sdk.NewCoin("nonntrn", math.NewInt(50))})

	feeKeeper.BurnAndDistribute(ctx)

	burnedAmount := feeKeeper.GetTotalBurnedNeutronsAmount(ctx)
	require.Equal(t, burnedAmount.Coin.Amount, math.NewInt(0))
}

func TestKeeper_BurnAndDistribute_SendCoinsFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	feeKeeper, ctx, mockBankKeeper, redistrAddr := setupBurnAndDistribute(t, ctrl, sdk.Coins{sdk.NewCoin("nonntrn", math.NewInt(50))})

	mockBankKeeper.EXPECT().SendCoins(ctx, redistrAddr, sdk.MustAccAddressFromBech32(feeKeeper.GetParams(ctx).TreasuryAddress), sdk.Coins{sdk.NewCoin("nonntrn", math.NewInt(50))}).Return(fmt.Errorf("testerror"))

	assert.Panics(t, func() {
		feeKeeper.BurnAndDistribute(ctx)
	}, "did not panic")

	burnedAmount := feeKeeper.GetTotalBurnedNeutronsAmount(ctx)
	require.Equal(t, burnedAmount.Coin.Amount, math.NewInt(0))
}

func TestKeeper_BurnAndDistribute_NtrnAndNonNtrn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	coins := sdk.Coins{sdk.NewCoin(feetypes.DefaultNeutronDenom, math.NewInt(70)), sdk.NewCoin("nonntrn", math.NewInt(20))}
	feeKeeper, ctx, mockBankKeeper, redistrAddr := setupBurnAndDistribute(t, ctrl, coins)

	mockBankKeeper.EXPECT().BurnCoins(ctx, revenuetypes.RevenueFeeRedistributePoolName, sdk.Coins{sdk.NewCoin(feetypes.DefaultNeutronDenom, math.NewInt(70))})
	mockBankKeeper.EXPECT().SendCoins(ctx, redistrAddr, sdk.MustAccAddressFromBech32(feeKeeper.GetParams(ctx).TreasuryAddress), sdk.Coins{sdk.NewCoin("nonntrn", math.NewInt(20))})

	feeKeeper.BurnAndDistribute(ctx)
	burnedAmount := feeKeeper.GetTotalBurnedNeutronsAmount(ctx)
	require.Equal(t, burnedAmount.Coin.Amount, math.NewInt(70))
}

func setupBurnAndDistribute(t *testing.T, ctrl *gomock.Controller, coins sdk.Coins) (*keeper.Keeper, sdk.Context, *mock_types.MockBankKeeper, sdk.AccAddress) {
	redistrAddr := sdk.AccAddress("neutronabcdasdf")
	mockAccountKeeper := mock_types.NewMockAccountKeeper(ctrl)
	mockBankKeeper := mock_types.NewMockBankKeeper(ctrl)
	feeKeeper, ctx := feekeeperutil.FeeburnerKeeperWithDeps(t, mockAccountKeeper, mockBankKeeper)

	mockAccountKeeper.EXPECT().GetModuleAddress(revenuetypes.RevenueFeeRedistributePoolName).Return(redistrAddr)
	mockBankKeeper.EXPECT().GetAllBalances(ctx, redistrAddr).Return(coins)

	return feeKeeper, ctx, mockBankKeeper, redistrAddr
}
