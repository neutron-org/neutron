package revenue_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/neutron-org/neutron/v5/testutil/revenue/keeper"
	"github.com/neutron-org/neutron/v5/x/revenue"
	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
	"github.com/stretchr/testify/require"
)

func TestInitAndExportGenesis(t *testing.T) {
	k, ctx := keeper.RevenueKeeper(t, nil, nil, "")

	// create some non-default genesis state
	genesisState := revenuetypes.DefaultGenesis()
	genesisState.Validators = append(genesisState.Validators, revenuetypes.ValidatorInfo{
		ConsensusAddress:            "neutronvalcons1arjwkww79m65csulawqngr7ngs4uqu5hv5736l",
		CommitedBlocksInPeriod:      1000,
		CommitedOracleVotesInPeriod: 1000,
	})
	genesisState.Params.PaymentScheduleType = revenuetypes.PAYMENT_SCHEDULE_TYPE_BLOCK_BASED
	ps := &revenuetypes.BlockBasedPaymentSchedule{BlocksPerPeriod: 5, CurrentPeriodStartBlock: 1}
	var err error
	genesisState.State.PaymentSchedule, err = codectypes.NewAnyWithValue(ps)
	require.Nil(t, err)

	// apply genesis state, export it, and compare
	revenue.InitGenesis(ctx, k, *genesisState)
	got := revenue.ExportGenesis(ctx, k)
	require.Equal(t, genesisState, got)
}

func TestGenesisSerialization(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	revenuetypes.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	genesisState := revenuetypes.DefaultGenesis()
	data, err := cdc.MarshalJSON(genesisState)
	require.NoError(t, err)
	err = genesisState.Validate()
	require.NoError(t, err)

	var genesisState2 revenuetypes.GenesisState
	err = cdc.UnmarshalJSON(data, &genesisState2)
	require.NoError(t, err)

	err = genesisState2.Validate()
	require.NoError(t, err)
}
