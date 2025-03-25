package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	dexmoduletypes "github.com/neutron-org/neutron/v6/x/dex/types"
)

func NewACoin(amt math.Int) sdk.Coin {
	return sdk.NewCoin("TokenA", amt)
}

func NewBCoin(amt math.Int) sdk.Coin {
	return sdk.NewCoin("TokenB", amt)
}

func FundAccount(bankKeeper bankkeeper.Keeper, ctx sdk.Context, addr sdk.AccAddress, amounts sdk.Coins) {
	if err := bankKeeper.MintCoins(ctx, dexmoduletypes.ModuleName, amounts); err != nil {
		panic(err)
	}

	if err := bankKeeper.SendCoinsFromModuleToAccount(ctx, dexmoduletypes.ModuleName, addr, amounts); err != nil {
		panic(err)
	}
}
