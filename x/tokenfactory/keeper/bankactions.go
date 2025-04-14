package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/neutron-org/neutron/v6/x/tokenfactory/types"
)

func (k Keeper) mintTo(ctx sdk.Context, amount sdk.Coin, mintTo string) error {
	// verify that denom is an x/tokenfactory denom
	_, _, err := types.DeconstructDenom(amount.Denom)
	if err != nil {
		return err
	}

	mintToAcc, err := sdk.AccAddressFromBech32(mintTo)
	if err != nil {
		return err
	}

	if k.isModuleAccount(ctx, mintToAcc) {
		return status.Errorf(codes.Internal, "minting to module accounts is forbidden")
	}

	if k.IsEscrowAddress(ctx, mintToAcc) {
		return status.Errorf(codes.Internal, "minting to IBC escrow accounts is forbidden")
	}

	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName,
		mintToAcc,
		sdk.NewCoins(amount))
}

func (k Keeper) burnFrom(ctx sdk.Context, amount sdk.Coin, burnFrom string) error {
	// verify that denom is an x/tokenfactory denom
	_, _, err := types.DeconstructDenom(amount.Denom)
	if err != nil {
		return err
	}

	burnFromAcc, err := sdk.AccAddressFromBech32(burnFrom)
	if err != nil {
		return err
	}

	if k.isModuleAccount(ctx, burnFromAcc) {
		return status.Errorf(codes.Internal, "burning from module accounts is forbidden")
	}

	if k.IsEscrowAddress(ctx, burnFromAcc) {
		return status.Errorf(codes.Internal, "burning from IBC escrow accounts is forbidden")
	}

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx,
		burnFromAcc,
		types.ModuleName,
		sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	return k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(amount))
}

func (k Keeper) forceTransfer(ctx sdk.Context, amount sdk.Coin, fromAddr, toAddr string) error {
	// verify that denom is an x/tokenfactory denom
	_, _, err := types.DeconstructDenom(amount.Denom)
	if err != nil {
		return err
	}

	transferFromAcc, err := sdk.AccAddressFromBech32(fromAddr)
	if err != nil {
		return err
	}

	transferToAcc, err := sdk.AccAddressFromBech32(toAddr)
	if err != nil {
		return err
	}

	if k.isModuleAccount(ctx, transferFromAcc) {
		return status.Errorf(codes.Internal, "force transfer from module accounts is forbidden")
	}

	if k.isModuleAccount(ctx, transferToAcc) {
		return status.Errorf(codes.Internal, "force transfer to module accounts is forbidden")
	}

	if k.IsEscrowAddress(ctx, transferFromAcc) {
		return status.Errorf(codes.Internal, "force transfer from IBC escrow accounts is forbidden")
	}

	if k.IsEscrowAddress(ctx, transferToAcc) {
		return status.Errorf(codes.Internal, "force transfer to IBC escrow accounts is forbidden")
	}

	return k.bankKeeper.SendCoins(ctx, transferFromAcc, transferToAcc, sdk.NewCoins(amount))
}

func (k Keeper) isModuleAccount(ctx sdk.Context, addr sdk.AccAddress) bool {
	for _, moduleName := range k.knownModules {
		account := k.accountKeeper.GetModuleAccount(ctx, moduleName)
		if account == nil {
			continue
		}

		if account.GetAddress().Equals(addr) {
			return true
		}
	}

	return false
}
