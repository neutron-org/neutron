package feerefunder_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/testutil/interchainqueries/nullify"
	"github.com/neutron-org/neutron/testutil/keeper"
	"github.com/neutron-org/neutron/x/feerefunder"
	"github.com/neutron-org/neutron/x/feerefunder/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		FeeInfos: []types.FeeInfo{types.FeeInfo{
			Payer:    "address",
			PacketId: types.NewPacketID("port", "channel", 64),
			Fee: types.Fee{
				RecvFee:    sdk.NewCoins(sdk.NewCoin("denom", sdk.NewInt(100))),
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom", sdk.NewInt(100))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom", sdk.NewInt(100))),
			},
		}},
	}

	require.EqualValues(t, genesisState.Params, types.DefaultParams())

	k, ctx := keeper.FeeKeeper(t)
	feerefunder.InitGenesis(ctx, *k, genesisState)
	got := feerefunder.ExportGenesis(ctx, *k)

	require.EqualValues(t, got.Params, types.DefaultParams())
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
