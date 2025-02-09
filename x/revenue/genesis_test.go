package revenue_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/neutron-org/neutron/v5/testutil/revenue/keeper"
	"github.com/neutron-org/neutron/v5/x/revenue"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
	"github.com/stretchr/testify/require"
)

func TestInitAndExportGenesis(t *testing.T) {
	k, ctx := keeper.RevenueKeeper(t, nil, nil, "")

	// create some non-default genesis state with all fields populated
	genesisState := revenuetypes.DefaultGenesis()
	genesisState.Validators = append(genesisState.Validators, revenuetypes.ValidatorInfo{
		ValOperAddress:              "neutronvaloper18zawa74y4xv6xg3zv0cstmfl9y38ecurgt4e70",
		CommitedBlocksInPeriod:      1000,
		CommitedOracleVotesInPeriod: 1000,
	})
	genesisState.Params.PaymentScheduleType = &revenuetypes.PaymentScheduleType{
		PaymentScheduleType: &revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{
			BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 5},
		},
	}
	ps := &revenuetypes.BlockBasedPaymentSchedule{BlocksPerPeriod: 5, CurrentPeriodStartBlock: 1}
	genesisState.PaymentSchedule = ps.IntoPaymentSchedule()
	genesisState.Prices = []*revenuetypes.RewardAssetPrice{
		{
			AbsolutePrice:   math.LegacyOneDec(),
			CumulativePrice: math.LegacyOneDec(),
			Timestamp:       1000,
		},
	}

	// apply genesis state, export it, and compare
	revenue.InitGenesis(ctx, k, *genesisState)
	got := revenue.ExportGenesis(ctx, k)
	require.Equal(t, genesisState, got)
}

func TestGenesisSerialization(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	revenuetypes.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	// create some non-default genesis state with all fields populated
	genesisState := revenuetypes.DefaultGenesis()
	genesisState.Validators = []revenuetypes.ValidatorInfo{
		{
			ValOperAddress:              "neutronvaloper1...",
			CommitedBlocksInPeriod:      100,
			CommitedOracleVotesInPeriod: 100,
		},
	}
	genesisState.Prices = []*revenuetypes.RewardAssetPrice{
		{
			AbsolutePrice:   math.LegacyOneDec(),
			CumulativePrice: math.LegacyOneDec(),
			Timestamp:       1000,
		},
	}

	data, err := cdc.MarshalJSON(genesisState)
	require.NoError(t, err)
	err = genesisState.Validate()
	require.NoError(t, err)

	var genesisState2 revenuetypes.GenesisState
	err = cdc.UnmarshalJSON(data, &genesisState2)
	require.NoError(t, err)

	err = genesisState2.Validate()
	require.NoError(t, err)
	require.Equal(t, genesisState.Params, genesisState2.Params)
	require.Equal(t, genesisState.Validators, genesisState2.Validators)
	require.Equal(t, genesisState.Prices, genesisState2.Prices)
	require.Equal(t, genesisState.PaymentSchedule, genesisState2.PaymentSchedule)
}
