package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/neutron-org/neutron/testutil/dex/keeper"
	"github.com/neutron-org/neutron/x/dex/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.DexKeeper(t)
	params := types.DefaultParams()

	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	require.EqualValues(t, params, k.GetParams(ctx))
}

func TestValidateParams(t *testing.T) {
	goodFees := []uint64{1, 2, 3, 4, 5, 200}
	require.NoError(t, types.Params{FeeTiers: goodFees}.Validate())

	badFees := []uint64{1, 2, 3, 3}
	require.Error(t, types.Params{FeeTiers: badFees}.Validate())
}
