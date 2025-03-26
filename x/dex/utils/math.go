package utils

import (
	"fmt"
	"math"
	"regexp"
	"strconv"

	sdkmath "cosmossdk.io/math"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
)

// Return the base value for price, 1.0001
func BasePrice() math_utils.PrecDec {
	return math_utils.MustNewPrecDecFromStr("1.0001")
}

func Abs(x int64) uint64 {
	if x < 0 {
		return uint64(-x)
	}

	return uint64(x)
}

func MinIntArr(vals []sdkmath.Int) sdkmath.Int {
	minInt := vals[0]
	for _, val := range vals {
		if val.LT(minInt) {
			minInt = val
		}
	}

	return minInt
}

func MaxIntArr(vals []sdkmath.Int) sdkmath.Int {
	maxInt := vals[0]
	for _, val := range vals {
		if val.GT(maxInt) {
			maxInt = val
		}
	}

	return maxInt
}

func Uint64ToSortableString(i uint64) string {
	// Converts a Uint to a string that sorts lexogrpahically in integer order
	intStr := strconv.FormatUint(i, 36)
	lenStr := len(intStr)
	lenChar := strconv.FormatUint(uint64(lenStr), 36)

	return fmt.Sprintf("%s%s", lenChar, intStr)
}

func SafeUint64ToInt64(in uint64) (out int64, overflow bool) {
	return int64(in), in > math.MaxInt64 //nolint:gosec
}

func MustSafeUint64ToInt64(in uint64) (out int64) {
	safeInt64, overflow := SafeUint64ToInt64(in)
	if overflow {
		panic("Overflow while casting uint64 to int64")
	}

	return safeInt64
}

var scientificNotationRE = regexp.MustCompile(`^([\d\.]+)([Ee]([\-\+])?(\d+))?$`)

func ParsePrecDecScientificNotation(n string) (math_utils.PrecDec, error) {
	match := scientificNotationRE.FindSubmatch([]byte(n))

	if len(match) == 0 {
		return math_utils.ZeroPrecDec(), fmt.Errorf("error parsing string as PrecDec with possible scientific notation")
	}

	baseDec, err := math_utils.NewPrecDecFromStr(string(match[1]))
	if err != nil {
		return math_utils.ZeroPrecDec(), fmt.Errorf("error parsing string as PrecDec with possible scientific notation: %v", err)
	}

	// No scientific notation
	if len(match[2]) == 0 {
		return baseDec, nil
	}

	pow, err := strconv.ParseUint(string(match[4]), 10, 64)
	if err != nil {
		return math_utils.ZeroPrecDec(), fmt.Errorf("error parsing string as PrecDec with possible scientific notation: %v", err)
	}

	shift := math_utils.NewPrecDec(10).Power(pow)

	if string(match[3]) == "-" { // negative exponent
		return baseDec.Quo(shift), nil
	} // else string(match[3]) == "+" || string(match[3]) == "" // positive exponent
	return baseDec.Mul(shift), nil
}
