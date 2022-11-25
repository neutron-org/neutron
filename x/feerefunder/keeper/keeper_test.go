package keeper_test

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	testutil_keeper "github.com/neutron-org/neutron/testutil/keeper"
	"github.com/pkg/errors"
	"testing"

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
			desc: "single proper denom but insufficient",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(1))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(1))),
			},
			valid: false,
		},
		{
			desc: "single denom sufficient amount",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(101))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(101))),
			},
			valid: true,
		},
		{
			desc: "multiple denoms, both are proper, only one enough",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(101)), sdk.NewCoin("denom2", sdk.NewInt(1))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(101)), sdk.NewCoin("denom2", sdk.NewInt(1))),
			},
			valid: true,
		},
		{
			desc: "no proper denom",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom3", sdk.NewInt(1))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom3", sdk.NewInt(1))),
			},
			valid: false,
		},
		{
			desc: "proper denom plus random one",
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
