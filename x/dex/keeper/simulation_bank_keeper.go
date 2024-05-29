package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v4/x/dex/types"
)

type SimulationBankKeeper struct {
	originalBankKeeper types.BankKeeper
}

func (s SimulationBankKeeper) SendCoinsFromAccountToModule(_ context.Context, _ sdk.AccAddress, _ string, _ sdk.Coins) error {
	return nil
}

func (s SimulationBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, _ string, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}

func (s SimulationBankKeeper) MintCoins(_ context.Context, _ string, _ sdk.Coins) error {
	return nil
}

func (s SimulationBankKeeper) BurnCoins(_ context.Context, _ string, _ sdk.Coins) error {
	return nil
}

func (s SimulationBankKeeper) IterateAccountBalances(ctx context.Context, addr sdk.AccAddress, cb func(sdk.Coin) bool) {
	s.originalBankKeeper.IterateAccountBalances(ctx, addr, cb)
}

func (s SimulationBankKeeper) GetSupply(ctx context.Context, denom string) sdk.Coin {
	return s.originalBankKeeper.GetSupply(ctx, denom)
}

func NewSimulationBankKeeper(bk types.BankKeeper) types.BankKeeper {
	return SimulationBankKeeper{originalBankKeeper: bk}
}
