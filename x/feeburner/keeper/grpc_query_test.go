package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"

	feekeeperutil "github.com/neutron-org/neutron/testutil/feeburner/keeper"
	"github.com/neutron-org/neutron/x/feeburner/types"
)

func TestGrpcQuery_TotalBurnedNeutronsAmount(t *testing.T) {
	feeKeeper, ctx := feekeeperutil.FeeburnerKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)

	feeKeeper.RecordBurnedFees(ctx, sdk.NewCoin(types.DefaultNeutronDenom, sdk.NewInt(100)))

	request := types.QueryTotalBurnedNeutronsAmountRequest{}
	response, err := feeKeeper.TotalBurnedNeutronsAmount(wctx, &request)
	require.NoError(t, err)
	require.Equal(t, &types.QueryTotalBurnedNeutronsAmountResponse{TotalBurnedNeutronsAmount: types.TotalBurnedNeutronsAmount{Coin: sdk.Coin{Denom: types.DefaultNeutronDenom, Amount: sdk.NewInt(100)}}}, response)
}
