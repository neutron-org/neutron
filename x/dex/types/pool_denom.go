package types

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	PoolDenomPrefix    = "neutron/pool/"
	PoolDenomRegexpStr = "^" + PoolDenomPrefix + `(\d+)` + "$"
)

var PoolDenomRegexp = regexp.MustCompile(PoolDenomRegexpStr)

func NewPoolDenom(poolID uint64) string {
	return fmt.Sprintf("%s%d", PoolDenomPrefix, poolID)
}

func ValidatePoolDenom(denom string) error {
	if _, err := ParsePoolIDFromDenom(denom); err != nil {
		return err
	}
	return nil
}

func ParsePoolIDFromDenom(denom string) (uint64, error) {
	res := PoolDenomRegexp.FindStringSubmatch(denom)
	if len(res) != 2 {
		return 0, ErrInvalidPoolDenom
	}
	idStr := res[1]

	idInt, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		return 0, ErrInvalidPoolDenom
	}

	return idInt, nil
}

// NewDexMintCoinsRestriction creates and returns a BankMintingRestrictionFn that only allows minting of
// valid pool denoms
func NewDexDenomMintCoinsRestriction() types.MintingRestrictionFn {
	return func(_ context.Context, coinsToMint sdk.Coins) error {
		for _, coin := range coinsToMint {
			err := ValidatePoolDenom(coin.Denom)
			if err != nil {
				return fmt.Errorf("does not have permission to mint %s", coin.Denom)
			}
		}
		return nil
	}
}
