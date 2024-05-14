package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/neutron-org/neutron/v4/testutil/dex/keeper"
	math_utils "github.com/neutron-org/neutron/v4/utils/math"
	"github.com/neutron-org/neutron/v4/x/dex/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.DexKeeper(t)
	params := types.DefaultParams()

	require.EqualValues(t, params, k.GetParams(ctx))
}

func TestSetParams(t *testing.T) {
	k, ctx := testkeeper.DexKeeper(t)

	newParams := types.Params{
		FeeTiers:              []uint64{0, 1},
		MaxTrueTakerSpread:    math_utils.MustNewPrecDecFromStr("0.111"),
		Max_JITsPerBlock:      0,
		GoodTilPurgeAllowance: 0,
	}
	err := k.SetParams(ctx, newParams)
	require.NoError(t, err)

	require.EqualValues(t, newParams, k.GetParams(ctx))
}

func TestValidateParams(t *testing.T) {
	goodFees := []uint64{1, 2, 3, 4, 5, 200}
	require.NoError(t, types.Params{FeeTiers: goodFees}.Validate())

	badFees := []uint64{1, 2, 3, 3}
	require.Error(t, types.Params{FeeTiers: badFees}.Validate())
}
