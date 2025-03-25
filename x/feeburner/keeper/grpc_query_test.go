package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	feekeeperutil "github.com/neutron-org/neutron/v6/testutil/feeburner/keeper"
	"github.com/neutron-org/neutron/v6/x/feeburner/types"
)

func TestGrpcQuery_TotalBurnedNeutronsAmount(t *testing.T) {
	feeKeeper, ctx := feekeeperutil.FeeburnerKeeper(t)

	feeKeeper.RecordBurnedFees(ctx, sdk.NewCoin(types.DefaultNeutronDenom, math.NewInt(100)))

	request := types.QueryTotalBurnedNeutronsAmountRequest{}
	response, err := feeKeeper.TotalBurnedNeutronsAmount(ctx, &request)
	require.NoError(t, err)
	require.Equal(t, &types.QueryTotalBurnedNeutronsAmountResponse{TotalBurnedNeutronsAmount: types.TotalBurnedNeutronsAmount{Coin: sdk.Coin{Denom: types.DefaultNeutronDenom, Amount: math.NewInt(100)}}}, response)
}
