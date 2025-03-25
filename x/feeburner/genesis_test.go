package feeburner_test

import (
	"testing"

	"cosmossdk.io/math"

	"github.com/neutron-org/neutron/v6/app/config"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/testutil/common/nullify"
	"github.com/neutron-org/neutron/v6/testutil/feeburner/keeper"
	"github.com/neutron-org/neutron/v6/x/feeburner"
	"github.com/neutron-org/neutron/v6/x/feeburner/types"
)

func TestGenesis(t *testing.T) {
	_ = config.GetDefaultConfig()

	amount := math.NewInt(10)

	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		TotalBurnedNeutronsAmount: types.TotalBurnedNeutronsAmount{
			Coin: sdk.NewCoin(types.DefaultNeutronDenom, amount),
		},
	}

	k, ctx := keeper.FeeburnerKeeper(t)
	feeburner.InitGenesis(ctx, *k, genesisState)

	burnedTokens := k.GetTotalBurnedNeutronsAmount(ctx)
	require.Equal(t, amount, burnedTokens.Coin.Amount)

	got := feeburner.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.Equal(t, amount, got.TotalBurnedNeutronsAmount.Coin.Amount)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
