package types

import (
	"bytes"
	_ "embed"
	"encoding/gob"
	fmt "fmt"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/utils"
)

const (
	// NOTE: 559_680 is the highest possible tick at which price can be calculated with a < 1% error
	// when using 27 digit decimal precision (via prec_dec).
	// The error rate for very negative ticks approaches zero, so there is no concern there
	MaxTickExp uint64 = 559_680
	MinPrice   string = "0.000000000000000000000000495"
	MaxPrice   string = "2020125331305056766452345.127500016657360222036663651"
)

//go:embed precomputed_prices.gob
var precomputedPricesBz []byte

var PrecomputedPrices []math_utils.PrecDec

func init() {
	err := loadPrecomputedPricesFromFile()
	if err != nil {
		panic(fmt.Sprintf("Failed to load precomputed powers: %v", err))
	}
}

func loadPrecomputedPricesFromFile() error {
	var stringPrices []string
	decoder := gob.NewDecoder(bytes.NewBuffer(precomputedPricesBz))
	err := decoder.Decode(&stringPrices)
	if err != nil {
		return err
	}

	// Convert the slice of strings back to math_utils.PrecDec
	PrecomputedPrices = make([]math_utils.PrecDec, len(stringPrices))
	for i, s := range stringPrices {
		PrecomputedPrices[i] = math_utils.MustNewPrecDecFromStr(s)
	}

	// Release precomputedPricesBz from memory
	precomputedPricesBz = []byte{}
	return nil
}

// Calculates the price for a swap from token 0 to token 1 given a relative tick
// tickIndex refers to the index of a specified tick such that x * 1.0001 ^(1 * t) = y
// Lower ticks offer better prices.
func CalcPrice(relativeTickIndex int64) (math_utils.PrecDec, error) {
	if IsTickOutOfRange(relativeTickIndex) {
		return math_utils.ZeroPrecDec(), ErrTickOutsideRange
	}
	if relativeTickIndex < 0 {
		return math_utils.OnePrecDec().Quo(PrecomputedPrices[-relativeTickIndex]), nil
	}
	// else
	return PrecomputedPrices[relativeTickIndex], nil
}

func BinarySearchPriceToTick(price math_utils.PrecDec) uint64 {
	if price.LT(math_utils.OnePrecDec()) {
		panic("Can only lookup prices <= 1")
	}
	var left uint64 // = 0
	right := MaxTickExp

	// Binary search to find the closest precomputed value
	for left < right {
		switch mid := (left + right) / 2; {
		case PrecomputedPrices[mid].Equal(price):
			return mid
		case PrecomputedPrices[mid].LT(price):
			left = mid + 1
		default:
			right = mid - 1

		}
	}

	// If exact match is not found, return the upper bound
	return right
}

func CalcTickIndexFromPrice(price math_utils.PrecDec) (int64, error) {
	if IsPriceOutOfRange(price) {
		return 0, ErrPriceOutsideRange
	}

	if price.LT(math_utils.OnePrecDec()) {
		// We only have a lookup table for prices >= 1
		// So we invert the price for the lookup
		invPrice := math_utils.OnePrecDec().Quo(price)
		tick := BinarySearchPriceToTick(invPrice)
		// flip the sign back the other direction
		return int64(tick) * -1, nil //nolint:gosec
	}

	tick := BinarySearchPriceToTick(price)

	return int64(tick), nil //nolint:gosec
}

func MustCalcPrice(relativeTickIndex int64) math_utils.PrecDec {
	price, err := CalcPrice(relativeTickIndex)
	if err != nil {
		panic(err)
	}
	return price
}

func IsTickOutOfRange(tickIndex int64) bool {
	return utils.Abs(tickIndex) > MaxTickExp
}

func IsPriceOutOfRange(price math_utils.PrecDec) bool {
	return price.GT(math_utils.MustNewPrecDecFromStr(MaxPrice)) ||
		price.LT(math_utils.MustNewPrecDecFromStr(MinPrice))
}

func ValidateTickFee(tick int64, fee uint64) error {
	// Ensure we do not overflow/wrap Uint
	if fee >= MaxTickExp {
		return ErrInvalidFee
	}
	// Ensure |tick| + fee <= MaxTickExp
	// NOTE: Ugly arithmetic is to ensure that we don't overflow uint64
	if utils.Abs(tick) > MaxTickExp-fee {
		return ErrTickOutsideRange
	}
	return nil
}

func ValidateFairOutput(amountIn math.Int, price math_utils.PrecDec) error {
	amountOut := math_utils.NewPrecDecFromInt(amountIn).Quo(price)
	if amountOut.LT(math_utils.OnePrecDec()) {
		return errors.Wrapf(ErrTradeTooSmall, "True output for %v tokens at price %v is %v", amountIn, price, amountOut)
	}
	return nil
}

// // Used for generating the precomputedPrice.gob file
// const PrecomputedPricesFile = "../types/precomputed_prices.gob"

// func generatePrecomputedPrices() []math_utils.PrecDec {
//	precomputedPowers := make([]math_utils.PrecDec, MaxTickExp+1)
//	precomputedPowers[0] = math_utils.OnePrecDec() // 1.0001^0 = 1
//	for i := 1; i <= int(MaxTickExp); i++ {
//		precomputedPowers[i] = precomputedPowers[i-1].Mul(utils.BasePrice())
//	}
//	return precomputedPowers
// }

// func WritePrecomputedPricesToFile() error {
//	computedPrices := generatePrecomputedPrices()
//	file, err := os.Create(PrecomputedPricesFile)
//	if err != nil {
//		panic(fmt.Sprintf("Error creating precomputed power file: %v", err.Error()))
//	}
//	defer file.Close()
//	stringPowers := make([]string, len(computedPrices))
//	for i, power := range computedPrices {
//		stringPowers[i] = power.String()
//	}

//	encoder := gob.NewEncoder(file)
//	err = encoder.Encode(stringPowers)
//	if err != nil {
//		panic(fmt.Sprintf("Error writing precomputed powers to file: %v", err.Error()))
//	}
//	return nil
// }
