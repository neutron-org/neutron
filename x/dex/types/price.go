package types

import (
	"bytes"
	_ "embed"
	"encoding/gob"
	fmt "fmt"
	"math/big"
	"os"

	math_utils "github.com/neutron-org/neutron/v8/utils/math"
	"github.com/neutron-org/neutron/v8/x/dex/utils"
)

const (
	// NOTE: 529_750 is the highest possible tick at which price can be calculated with a < 1% error and all prices are unique for negative ticks
	// when using 27 digit decimal precision (via prec_dec).
	// The error rate for very negative ticks approaches zero, so there is no concern there
	MaxTickExp       uint64 = 529_750
	MinPrice         string = "0.000000000000000000000009871"
	MaxPrice         string = "101297777749006516066611.914775584130706898691360168"
	PriceArrayOffset int64  = int64(MaxTickExp)
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
	// else
	return PrecomputedPrices[relativeTickIndex+PriceArrayOffset], nil
}

func BinarySearchPriceToTick(price math_utils.PrecDec) int64 {
	var left int64 = 0
	right := int64(MaxTickExp) * 2

	// Binary search to find the closest precomputed value
	for left < right {
		switch mid := (left + right) / 2; {
		case PrecomputedPrices[mid].Equal(price):
			return mid - PriceArrayOffset
		case PrecomputedPrices[mid].LT(price):
			left = mid + 1
		default:
			right = mid - 1

		}
	}

	// If exact match is not found, return the upper bound
	return right - PriceArrayOffset
}

func CalcTickIndexFromPrice(price math_utils.PrecDec) (int64, error) {
	if IsPriceOutOfRange(price) {
		return 0, ErrPriceOutsideRange
	}

	tick := BinarySearchPriceToTick(price)

	return tick, nil
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

// Used for generating the precomputedPrice.gob file
const PrecomputedPricesFile = "./x/dex/types/precomputed_prices.gob"

func BigFloatToPrecDec(f *big.Float) math_utils.PrecDec {
	var tenPow27 = new(big.Int).Exp(big.NewInt(10), big.NewInt(math_utils.Precision), nil)
	prec := f.Prec()

	// scaled = f * 10^27 (exact scale), with rounding to nearest
	scaled := new(big.Float).SetPrec(prec)
	scaled.Mul(f, new(big.Float).SetInt(tenPow27))
	i, _ := scaled.Int(nil)
	return math_utils.NewPrecDecFromBigIntWithPrec(i, math_utils.Precision)
}

// Calculate all tick prices using Big.Float with 256 precision and then convert to PrecDec
func generatePrecomputedPrices() []math_utils.PrecDec {

	precomputedPowers := make([]math_utils.PrecDec, MaxTickExp*2+1)
	precomputedPowers[0] = math_utils.OnePrecDec() // 1.0001^0 = 1
	base := big.NewFloat(1.0001).SetPrec(256)
	prevVal := big.NewFloat(1).SetPrec(256)
	// Generate all positive tick prices
	for i := 1; i <= int(MaxTickExp); i++ {
		nextVal := new(big.Float).Mul(prevVal, base)
		precomputedPowers[int64(i)+PriceArrayOffset] = BigFloatToPrecDec(nextVal)
		prevVal = nextVal
	}
	precomputedPowers[PriceArrayOffset] = math_utils.OnePrecDec()
	// Generate all negative tick prices
	prevVal = big.NewFloat(1).SetPrec(256)
	for i := int64(-1); i >= int64(MaxTickExp)*-1; i-- {
		nextVal := new(big.Float).Quo(prevVal, base)
		precomputedPowers[int64(i)+PriceArrayOffset] = BigFloatToPrecDec(nextVal)
		fmt.Println(i, BigFloatToPrecDec(nextVal))
		prevVal = nextVal
		fmt.Println(i, prevVal)

	}
	return precomputedPowers
}

func WritePrecomputedPricesToFile() error {
	computedPrices := generatePrecomputedPrices()
	file, err := os.Create(PrecomputedPricesFile)
	if err != nil {
		panic(fmt.Sprintf("Error creating precomputed power file: %v", err.Error()))
	}
	defer file.Close()
	stringPowers := make([]string, len(computedPrices))
	for i, power := range computedPrices {
		stringPowers[i] = power.String()
	}

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(stringPowers)
	if err != nil {
		panic(fmt.Sprintf("Error writing precomputed powers to file: %v", err.Error()))
	}
	return nil
}
