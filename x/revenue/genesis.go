package revenue

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"

	"github.com/neutron-org/neutron/v6/x/revenue/keeper"
	"github.com/neutron-org/neutron/v6/x/revenue/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k *keeper.Keeper, genState types.GenesisState) {
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	if genState.PaymentSchedule == nil {
		genState.PaymentSchedule = types.PaymentScheduleIByType(
			genState.Params.PaymentScheduleType.PaymentScheduleType,
		).IntoPaymentSchedule()
	}
	if err := k.SetPaymentSchedule(ctx, genState.PaymentSchedule); err != nil {
		panic(err)
	}

	ps, err := genState.PaymentSchedule.IntoPaymentScheduleI()
	if err != nil {
		panic(err)
	}

	blockHeight := uint64(ctx.BlockHeight()) //nolint:gosec
	blocksInPeriod := ps.TotalBlocksInPeriod(ctx)

	var currentPeriodStartBlock uint64

	switch v := ps.(type) {
	case *types.BlockBasedPaymentSchedule:
		currentPeriodStartBlock = v.CurrentPeriodStartBlock
	case *types.MonthlyPaymentSchedule:
		currentPeriodStartBlock = v.CurrentMonthStartBlock
	}

	for _, elem := range genState.Validators {
		_, addr, err := bech32.DecodeAndConvert(elem.ValOperAddress)
		if err != nil {
			panic(err)
		}

		if blockHeight == 0 {
			if elem.CommitedBlocksInPeriod > 0 {
				panic(fmt.Sprintf("Non-zero CommitedBlocksInPeriod for validator %s", elem.ValOperAddress))
			}

			if elem.CommitedOracleVotesInPeriod > 0 {
				panic(fmt.Sprintf("Non-zero CommitedOracleVotesInPeriod for validator %s", elem.ValOperAddress))
			}
		} else {
			if currentPeriodStartBlock > blockHeight {
				panic("currentPeriodStartBlock exceeds current block height")
			}

			if elem.CommitedBlocksInPeriod > blocksInPeriod {
				panic(fmt.Sprintf("CommitedBlocksInPeriod exceeds the initial payment schedule block period for validator %s", elem.ValOperAddress))
			}

			if elem.CommitedOracleVotesInPeriod > blocksInPeriod {
				panic(fmt.Sprintf("CommitedOracleVotesInPeriod exceeds the initial payment schedule block period for validator %s", elem.ValOperAddress))
			}
		}

		if err := k.SetValidatorInfo(ctx, addr, elem); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	var err error
	genesis := types.DefaultGenesis()
	genesis.Params, err = k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	ps, err := k.GetPaymentScheduleI(ctx)
	if err != nil {
		panic(err)
	}
	genesis.PaymentSchedule = ps.IntoPaymentSchedule()

	genesis.Validators, err = k.GetAllValidatorInfo(ctx)
	if err != nil {
		panic(err)
	}

	return genesis
}
