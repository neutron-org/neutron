package keeper

import (
	"fmt"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/neutron-org/neutron/x/tokenfactory/types"
)

// CreateDenom creates new denom with `factory/{creatorAddr}/{subdenom}` name.
// Charges creatorAddr fee for creation
func (k Keeper) CreateDenom(ctx sdk.Context, creatorAddr, subdenom string) (newTokenDenom string, err error) {
	err = k.chargeFeeForDenomCreation(ctx, creatorAddr)
	if err != nil {
		return "", errors.Wrapf(types.ErrUnableToCharge, "denom fee collection error: %v", err)
	}

	denom, err := k.validateCreateDenom(ctx, creatorAddr, subdenom)
	if err != nil {
		return "", errors.Wrapf(types.ErrInvalidDenom, "denom validation error: %v", err)
	}

	err = k.createDenomAfterValidation(ctx, creatorAddr, denom)
	if err != nil {
		return "", errors.Wrap(err, "create denom after validation error")
	}

	return denom, nil
}

// Runs CreateDenom logic after the charge and all denom validation has been handled.
// Made into a second function for genesis initialization.
func (k Keeper) createDenomAfterValidation(ctx sdk.Context, creatorAddr, denom string) (err error) {
	denomMetaData := banktypes.Metadata{
		DenomUnits: []*banktypes.DenomUnit{{
			Denom:    denom,
			Exponent: 0,
		}},
		Base: denom,
	}

	k.bankKeeper.SetDenomMetaData(ctx, denomMetaData)

	authorityMetadata := types.DenomAuthorityMetadata{
		Admin: creatorAddr,
	}
	err = k.setAuthorityMetadata(ctx, denom, authorityMetadata)
	if err != nil {
		return errors.Wrapf(types.ErrInvalidAuthorityMetadata, "unable to set authority metadata: %v", err)
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
		return "", errors.Wrapf(types.ErrTokenDenom, "wrong denom token: %v", err)
	}

	_, found := k.bankKeeper.GetDenomMetaData(ctx, denom)
	if found {
		return "", types.ErrDenomExists
	}

	return denom, nil
}

func (k Keeper) chargeFeeForDenomCreation(ctx sdk.Context, creatorAddr string) (err error) {
	// Send creation fee to community pool
	creationFee := k.GetParams(ctx).DenomCreationFee
	accAddr, err := sdk.AccAddressFromBech32(creatorAddr)
	if err != nil {
		return errors.Wrapf(types.ErrUnableToCharge, "wrong creator address: %v", err)
	}

	params := k.GetParams(ctx)

	if len(creationFee) > 0 {
		feeCollectorAddr, err := sdk.AccAddressFromBech32(params.FeeCollectorAddress)
		if err != nil {
			return errors.Wrapf(types.ErrUnableToCharge, "wrong fee collector address: %v", err)
		}

		err = k.bankKeeper.SendCoins(
			ctx,
			accAddr, feeCollectorAddr,
			creationFee,
		)

		if err != nil {
			return errors.Wrap(err, "unable to send coins to fee collector")
		}
	}

	return nil
}
