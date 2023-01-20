package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"testing"

	feekeeperutil "github.com/neutron-org/neutron/testutil/feeburner/keeper"
	feetypes "github.com/neutron-org/neutron/x/feeburner/types"
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
			amount:    sdk.Coin{Denom: feetypes.DefaultNeutronDenom, Amount: sdk.NewInt(300)},
			expected:  &sdk.Coin{Denom: feetypes.DefaultNeutronDenom, Amount: sdk.NewInt(300)},
			wantPanic: false,
		},
		{
			desc:      "default works with existing tokens burned sums",
			initial:   &sdk.Coin{Denom: feetypes.DefaultNeutronDenom, Amount: sdk.NewInt(100)},
			amount:    sdk.Coin{Denom: feetypes.DefaultNeutronDenom, Amount: sdk.NewInt(300)},
			expected:  &sdk.Coin{Denom: feetypes.DefaultNeutronDenom, Amount: sdk.NewInt(400)},
			wantPanic: false,
		},
		{
			desc:      "with non-default denom should not write it to the store",
			initial:   nil,
			amount:    sdk.Coin{Denom: "nondefaultdenom", Amount: sdk.NewInt(300)},
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
			expected: &sdk.Coin{Denom: feetypes.DefaultNeutronDenom, Amount: sdk.NewInt(0)},
		},
		{
			desc:     "works with existing burned fees",
			initial:  &sdk.Coin{Denom: feetypes.DefaultNeutronDenom, Amount: sdk.NewInt(100)},
			expected: &sdk.Coin{Denom: feetypes.DefaultNeutronDenom, Amount: sdk.NewInt(100)},
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

// Test BurnAndDistributeFunds
// 1. Nothing to burn and distribute
// 2. Has NTRN tokens to burn
// 3. Has non-NTRN tokens to distribute
