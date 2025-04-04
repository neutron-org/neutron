package utils

import (
	"strconv"
	"strings"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	base   = 10
	bitlen = 64
)

func ParseUint64SliceFromString(s, separator string) ([]uint64, error) {
	var parsedInts []uint64
	for _, s := range strings.Split(s, separator) {
		s = strings.TrimSpace(s)

		parsed, err := strconv.ParseUint(s, base, bitlen)
		if err != nil {
			return []uint64{}, err
		}
		parsedInts = append(parsedInts, parsed)
	}
	return parsedInts, nil
}

func ParseSdkIntFromString(s, separator string) ([]math.Int, error) {
	var parsedInts []math.Int
	for _, weightStr := range strings.Split(s, separator) {
		weightStr = strings.TrimSpace(weightStr)

		parsed, err := strconv.ParseUint(weightStr, base, bitlen)
		if err != nil {
			return parsedInts, err
		}
		parsedInts = append(parsedInts, math.NewIntFromUint64(parsed))
	}
	return parsedInts, nil
}

func ParseSdkDecFromString(s, separator string) ([]math.LegacyDec, error) {
	var parsedDec []math.LegacyDec
	for _, weightStr := range strings.Split(s, separator) {
		weightStr = strings.TrimSpace(weightStr)

		parsed, err := math.LegacyNewDecFromStr(weightStr)
		if err != nil {
			return parsedDec, err
		}

		parsedDec = append(parsedDec, parsed)
	}
	return parsedDec, nil
}

// CreateRandomAccounts is a function that returns a list of randomly generated AccAddresses
func CreateRandomAccounts(numAccts int) []sdk.AccAddress {
	testAddrs := make([]sdk.AccAddress, numAccts)
	for i := 0; i < numAccts; i++ {
		pk := ed25519.GenPrivKey().PubKey()
		testAddrs[i] = sdk.AccAddress(pk.Address())
	}

	return testAddrs
}
