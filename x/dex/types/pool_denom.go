package types

import (
	"fmt"
	"regexp"
	"strconv"
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
	if !PoolDenomRegexp.MatchString(denom) {
		return ErrInvalidPoolDenom
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
