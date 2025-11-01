package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleDenomPrefix = "coinfactory"
	// See the CoinFactory readme for a derivation of these.
	// TL;DR, MaxSubdenomLength + MaxHrpLength = 60 comes from SDK max denom length = 128
	// and the structure of coinfactory denoms.
	MaxSubdenomLength = 44
	MaxHrpLength      = 16
	// MaxCreatorLength = 59 + MaxHrpLength
	MaxCreatorLength = 59 + MaxHrpLength
	Separator        = "."
)

// GetTokenDenom constructs a denom string for tokens created by coinfactory
// based on an input creator address and a subdenom
// The denom constructed is coinfactory.{creator}.{subdenom}
func GetTokenDenom(creator, subdenom string) (string, error) {
	if len(subdenom) > MaxSubdenomLength {
		return "", ErrSubdenomTooLong
	}
	if len(creator) > MaxCreatorLength {
		return "", ErrCreatorTooLong
	}
	if strings.Contains(creator, Separator) {
		return "", ErrInvalidCreator
	}
	denom := strings.Join([]string{ModuleDenomPrefix, creator, subdenom}, Separator)
	return denom, sdk.ValidateDenom(denom)
}

// DeconstructDenom takes a token denom string and verifies that it is a valid
// denom of the coinfactory module, and is of the form `coinfactory + Separator + {creator} + Separator + {subdenom}`
// If valid, it returns the creator address and subdenom
func DeconstructDenom(denom string) (creator, subdenom string, err error) {
	err = sdk.ValidateDenom(denom)
	if err != nil {
		return "", "", err
	}

	strParts := strings.Split(denom, Separator)
	if len(strParts) < 3 {
		return "", "", errorsmod.Wrapf(ErrInvalidDenom, "not enough parts of denom %s", denom)
	}

	if strParts[0] != ModuleDenomPrefix {
		return "", "", errorsmod.Wrapf(ErrInvalidDenom, "denom prefix is incorrect. Is: %s.  Should be: %s", strParts[0], ModuleDenomPrefix)
	}

	creator = strParts[1]
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	if err != nil {
		return "", "", errorsmod.Wrapf(ErrInvalidDenom, "Invalid creator address (%s)", err)
	}

	// Handle the case where a denom has a Separator in its subdenom. For example,
	// when we did the split, we'd turn factory + Separator + accaddr + Separator + atomderivative + Separator + sikka into ["factory", "accaddr", "atomderivative", "sikka"]
	// So we have to join [2:] with a Separator as the delimiter to get back the correct subdenom which should be "atomderivative + Separator + sikka"
	subdenom = strings.Join(strParts[2:], Separator)

	return creatorAddr.String(), subdenom, nil
}
