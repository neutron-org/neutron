package revenue

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"

	"github.com/neutron-org/neutron/v5/x/revenue/keeper"
	"github.com/neutron-org/neutron/v5/x/revenue/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k *keeper.Keeper, genState types.GenesisState) {
	for _, elem := range genState.Validators {
		_, addr, err := bech32.DecodeAndConvert(elem.ConsensusAddress)
		if err != nil {
			panic(err)
		}

		err = k.SetValidatorInfo(ctx, addr, elem)
		if err != nil {
			panic(err)
		}
	}

	// TODO: test export import genesis
	for _, elem := range genState.CumulativePrices {
		err := k.SaveCumulativePrice(ctx, elem.LastPrice, elem.Timestamp)
		if err != nil {
			panic(err)
		}
	}

	err := k.SetParams(ctx, genState.Params)
	if err != nil {
		panic(err)
	}

	err = k.SetState(ctx, genState.State)
	if err != nil {
		panic(err)
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
	genesis.State, err = k.GetState(ctx)
	if err != nil {
		panic(err)
	}

	genesis.Validators, err = k.GetAllValidatorInfo(ctx)
	if err != nil {
		panic(err)
	}

	genesis.CumulativePrices, err = k.GetAllCumulativePrices(ctx)
	if err != nil {
		panic(err)
	}

	return genesis
}
