package revenue_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/testutil/revenue/keeper"
	"github.com/neutron-org/neutron/v6/x/revenue"
	revenuetypes "github.com/neutron-org/neutron/v6/x/revenue/types"
)

func TestInitAndExportGenesis(t *testing.T) {
	k, ctx := keeper.RevenueKeeper(t, nil, nil, "")

	// create some non-default genesis state with all fields populated
	genesisState := revenuetypes.DefaultGenesis()
	genesisState.Validators = append(genesisState.Validators, revenuetypes.ValidatorInfo{
		ValOperAddress:              val1OperAddr,
		CommitedBlocksInPeriod:      0,
		CommitedOracleVotesInPeriod: 0,
	})
	genesisState.Params.PaymentScheduleType = &revenuetypes.PaymentScheduleType{
		PaymentScheduleType: &revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{
			BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 5},
		},
	}
	ps := &revenuetypes.BlockBasedPaymentSchedule{BlocksPerPeriod: 5, CurrentPeriodStartBlock: 1}
	genesisState.PaymentSchedule = ps.IntoPaymentSchedule()

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
	require.Equal(t, genesisState.PaymentSchedule, genesisState2.PaymentSchedule)
}

func TestGenesisInvalidCommitedBlocksInPeriodForZeroHeight(t *testing.T) {
	k, ctx := keeper.RevenueKeeper(t, nil, nil, "")

	valOperAddress := val1OperAddr

	genesisState := revenuetypes.DefaultGenesis()
	genesisState.Validators = append(genesisState.Validators, revenuetypes.ValidatorInfo{
		ValOperAddress:              valOperAddress,
		CommitedBlocksInPeriod:      1000,
		CommitedOracleVotesInPeriod: 0,
	})
	genesisState.Params.PaymentScheduleType = &revenuetypes.PaymentScheduleType{
		PaymentScheduleType: &revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{
			BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 5},
		},
	}
	ps := &revenuetypes.BlockBasedPaymentSchedule{BlocksPerPeriod: 5, CurrentPeriodStartBlock: 1}
	genesisState.PaymentSchedule = ps.IntoPaymentSchedule()

	require.PanicsWithValue(t, fmt.Sprintf("Non-zero CommitedBlocksInPeriod for validator %s", valOperAddress), func() { revenue.InitGenesis(ctx, k, *genesisState) })
}

func TestGenesisInvalidCommitedOracleVotesInPeriodForZeroHeight(t *testing.T) {
	k, ctx := keeper.RevenueKeeper(t, nil, nil, "")

	valOperAddress := val1OperAddr

	genesisState := revenuetypes.DefaultGenesis()
	genesisState.Validators = append(genesisState.Validators, revenuetypes.ValidatorInfo{
		ValOperAddress:              valOperAddress,
		CommitedBlocksInPeriod:      0,
		CommitedOracleVotesInPeriod: 1000,
	})
	genesisState.Params.PaymentScheduleType = &revenuetypes.PaymentScheduleType{
		PaymentScheduleType: &revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{
			BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 5},
		},
	}
	ps := &revenuetypes.BlockBasedPaymentSchedule{BlocksPerPeriod: 5, CurrentPeriodStartBlock: 1}
	genesisState.PaymentSchedule = ps.IntoPaymentSchedule()

	require.PanicsWithValue(t, fmt.Sprintf("Non-zero CommitedOracleVotesInPeriod for validator %s", valOperAddress), func() { revenue.InitGenesis(ctx, k, *genesisState) })
}

func TestGenesisInvalidCurrentPeriodStartBlock(t *testing.T) {
	k, ctx := keeper.RevenueKeeper(t, nil, nil, "")
	ctx = ctx.WithBlockHeight(2)

	valOperAddress := val1OperAddr

	genesisState := revenuetypes.DefaultGenesis()
	genesisState.Validators = append(genesisState.Validators, revenuetypes.ValidatorInfo{
		ValOperAddress:              valOperAddress,
		CommitedBlocksInPeriod:      1000,
		CommitedOracleVotesInPeriod: 1000,
	})
	genesisState.Params.PaymentScheduleType = &revenuetypes.PaymentScheduleType{
		PaymentScheduleType: &revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{
			BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 5},
		},
	}
	ps := &revenuetypes.BlockBasedPaymentSchedule{BlocksPerPeriod: 5, CurrentPeriodStartBlock: 3}
	genesisState.PaymentSchedule = ps.IntoPaymentSchedule()

	require.PanicsWithValue(t, "currentPeriodStartBlock exceeds current block height", func() { revenue.InitGenesis(ctx, k, *genesisState) })
}

func TestGenesisInvalidCommitedBlocksInPeriodForNonZeroHeight(t *testing.T) {
	k, ctx := keeper.RevenueKeeper(t, nil, nil, "")
	ctx = ctx.WithBlockHeight(2)

	valOperAddress := val1OperAddr

	genesisState := revenuetypes.DefaultGenesis()
	genesisState.Validators = append(genesisState.Validators, revenuetypes.ValidatorInfo{
		ValOperAddress:              valOperAddress,
		CommitedBlocksInPeriod:      1000,
		CommitedOracleVotesInPeriod: 0,
	})
	genesisState.Params.PaymentScheduleType = &revenuetypes.PaymentScheduleType{
		PaymentScheduleType: &revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{
			BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 5},
		},
	}
	ps := &revenuetypes.BlockBasedPaymentSchedule{BlocksPerPeriod: 5, CurrentPeriodStartBlock: 1}
	genesisState.PaymentSchedule = ps.IntoPaymentSchedule()

	require.PanicsWithValue(t, fmt.Sprintf("CommitedBlocksInPeriod exceeds the initial payment schedule block period for validator %s", valOperAddress), func() { revenue.InitGenesis(ctx, k, *genesisState) })
}

func TestGenesisInvalidCommitedOracleVotesInPeriodForNonZeroHeight(t *testing.T) {
	k, ctx := keeper.RevenueKeeper(t, nil, nil, "")
	ctx = ctx.WithBlockHeight(2)

	valOperAddress := val1OperAddr

	genesisState := revenuetypes.DefaultGenesis()
	genesisState.Validators = append(genesisState.Validators, revenuetypes.ValidatorInfo{
		ValOperAddress:              valOperAddress,
		CommitedBlocksInPeriod:      0,
		CommitedOracleVotesInPeriod: 1000,
	})
	genesisState.Params.PaymentScheduleType = &revenuetypes.PaymentScheduleType{
		PaymentScheduleType: &revenuetypes.PaymentScheduleType_BlockBasedPaymentScheduleType{
			BlockBasedPaymentScheduleType: &revenuetypes.BlockBasedPaymentScheduleType{BlocksPerPeriod: 5},
		},
	}
	ps := &revenuetypes.BlockBasedPaymentSchedule{BlocksPerPeriod: 5, CurrentPeriodStartBlock: 1}
	genesisState.PaymentSchedule = ps.IntoPaymentSchedule()

	require.PanicsWithValue(t, fmt.Sprintf("CommitedOracleVotesInPeriod exceeds the initial payment schedule block period for validator %s", valOperAddress), func() { revenue.InitGenesis(ctx, k, *genesisState) })
}
