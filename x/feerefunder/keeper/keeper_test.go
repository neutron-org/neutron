package keeper_test

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	testutil_keeper "github.com/neutron-org/neutron/testutil/keeper"
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/x/feerefunder/types"
	"github.com/stretchr/testify/require"
)

func TestKeeperCheckFees(t *testing.T) {
	k, ctx := testutil_keeper.FeeKeeper(t)

	k.SetParams(ctx, types.Params{
		MinFee: types.Fee{
			RecvFee:    nil,
			AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(100)), sdk.NewCoin("denom2", sdk.NewInt(100))),
			TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(100)), sdk.NewCoin("denom2", sdk.NewInt(100))),
		},
	})

	for _, tc := range []struct {
		desc  string
		fees  *types.Fee
		valid bool
	}{
		{
			desc: "SingleProperDenomInsufficient",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(1))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(1))),
			},
			valid: false,
		},
		{
			desc: "SingleDenomSufficient",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(101))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(101))),
			},
			valid: true,
		},
		{
			desc: "MultipleDenomsOneIsEnough",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(101)), sdk.NewCoin("denom2", sdk.NewInt(1))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(101)), sdk.NewCoin("denom2", sdk.NewInt(1))),
			},
			valid: true,
		},
		{
			desc: "NoProperDenom",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom3", sdk.NewInt(1))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom3", sdk.NewInt(1))),
			},
			valid: false,
		},
		{
			desc: "ProperDenomPlusRandomOne",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(101)), sdk.NewCoin("denom3", sdk.NewInt(1))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(101)), sdk.NewCoin("denom3", sdk.NewInt(1))),
			},
			valid: true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := k.CheckFees(ctx, *tc.fees)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.IsType(t, errors.WithStack(sdkerrors.ErrInsufficientFee), errors.Unwrap(err))
			}
		})
	}
}
