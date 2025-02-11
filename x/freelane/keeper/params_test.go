package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/neutron-org/neutron/v5/testutil/freelane/keeper"
	"github.com/neutron-org/neutron/v5/x/freelane/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.FreeLaneKeeper(t)
	params := types.Params{
		BlockSpace:     0.07,
		SequenceNumber: 4,
	}

	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	require.EqualValues(t, params, k.GetParams(ctx))
}
