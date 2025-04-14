package keeper

import (
	"fmt"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/neutron-org/neutron/v6/x/tokenfactory/types"
)

// ConvertToBaseToken converts a fee amount in a whitelisted fee token to the base fee token amount
func (k Keeper) CreateDenom(ctx sdk.Context, creatorAddr, subdenom string) (newTokenDenom string, err error) {
	denom, err := k.validateCreateDenom(ctx, creatorAddr, subdenom)
	if err != nil {
		return "", err
	}

	err = k.chargeForCreateDenom(ctx, creatorAddr)
	if err != nil {
		return "", err
	}

	err = k.createDenomAfterValidation(ctx, creatorAddr, denom)
	return denom, err
}

// Runs CreateDenom logic after the charge and all denom validation has been handled.
// Made into a second function for genesis initialization.
func (k Keeper) createDenomAfterValidation(ctx sdk.Context, creatorAddr, denom string) (err error) {
	_, exists := k.bankKeeper.GetDenomMetaData(ctx, denom)
	if !exists {
		denomMetaData := banktypes.Metadata{
			DenomUnits: []*banktypes.DenomUnit{{
				Denom:    denom,
				Exponent: 0,
			}},
			Base: denom,
		}

		k.bankKeeper.SetDenomMetaData(ctx, denomMetaData)
	}

	authorityMetadata := types.DenomAuthorityMetadata{
		Admin: creatorAddr,
	}
	err = k.setAuthorityMetadata(ctx, denom, authorityMetadata)
	if err != nil {
		return err
	}

	k.addDenomFromCreator(ctx, creatorAddr, denom)
	return nil
}

func (k Keeper) validateCreateDenom(ctx sdk.Context, creatorAddr, subdenom string) (newTokenDenom string, err error) {
	// Temporary check until IBC bug is sorted out
	if k.bankKeeper.HasSupply(ctx, subdenom) {
		return "", fmt.Errorf("temporary error until IBC bug is sorted out, " +
			"can't create subdenoms that are the same as a native denom")
	}

	denom, err := types.GetTokenDenom(creatorAddr, subdenom)
	if err != nil {
		return "", err
	}

	_, found := k.bankKeeper.GetDenomMetaData(ctx, denom)
	if found {
		return "", types.ErrDenomExists
	}

	return denom, nil
}

func (k Keeper) chargeForCreateDenom(ctx sdk.Context, creatorAddr string) (err error) {
	params := k.GetParams(ctx)

	// ORIGINAL: if DenomCreationFee is non-zero, transfer the tokens from the creator
	// account to community pool
	// MODIFIED: if DenomCreationFee is non-zero, transfer the tokens from the creator
	// account to feeCollectorAddr
	if len(params.DenomCreationFee) != 0 {
		accAddr, err := sdk.AccAddressFromBech32(creatorAddr)
		if err != nil {
			return err
		}
		// Instead of funding community pool we send funds to fee collector addr
		// if err := k.communityPoolKeeper.FundCommunityPool(ctx, params.DenomCreationFee, accAddr); err != nil {
		//	return err
		//}

		feeCollectorAddr, err := sdk.AccAddressFromBech32(params.FeeCollectorAddress)
		if err != nil {
			return errors.Wrapf(err, "wrong fee collector address: %v", err)
		}

		err = k.bankKeeper.SendCoins(
			ctx,
			accAddr, feeCollectorAddr,
			params.DenomCreationFee,
		)
		if err != nil {
			return errors.Wrap(err, "unable to send coins to fee collector")
		}
	}

	// if DenomCreationGasConsume is non-zero, consume the gas
	if params.DenomCreationGasConsume != 0 {
		ctx.GasMeter().ConsumeGas(params.DenomCreationGasConsume, "consume denom creation gas")
	}

	return nil
}
