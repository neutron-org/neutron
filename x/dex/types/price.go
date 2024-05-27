package types

import (
	"cosmossdk.io/errors"
	"cosmossdk.io/math"

	math_utils "github.com/neutron-org/neutron/v4/utils/math"
	"github.com/neutron-org/neutron/v4/x/dex/utils"
)

// NOTE: 559_680 is the highest possible tick at which price can be calculated with a < 1% error
// when using 26 digit decimal precision (via prec_dec).
// The error rate for very negative ticks approaches zero, so there is no concern there
const MaxTickExp uint64 = 559_680

// Calculates the price for a swap from token 0 to token 1 given a relative tick
// tickIndex refers to the index of a specified tick such that x * 1.0001 ^(-1 * t) = y
// Lower ticks offer better prices.
func CalcPrice(relativeTickIndex int64) (math_utils.PrecDec, error) {
	if IsTickOutOfRange(relativeTickIndex) {
		return math_utils.ZeroPrecDec(), ErrTickOutsideRange
	}
	if relativeTickIndex < 0 {
		return utils.BasePrice().Power(uint64(-1 * relativeTickIndex)), nil
	}
	// else
	return math_utils.OnePrecDec().Quo(utils.BasePrice().Power(uint64(relativeTickIndex))), nil
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

func ValidateTickFee(tick int64, fee uint64) error {
	// Ensure |tick| + fee <= MaxTickExp
	// NOTE: Ugly arithmetic is to ensure that we don't overflow uint64
	if utils.Abs(tick) > MaxTickExp-fee {
		return ErrTickOutsideRange
	}
	return nil
}

func ValidateFairOutput(amountIn math.Int, price math_utils.PrecDec) error {
	amountOut := price.MulInt(amountIn)
	if amountOut.LT(math_utils.OnePrecDec()) {
		return errors.Wrapf(ErrTradeTooSmall, "True output for %v tokens at price %v is %v", amountIn, price, amountOut)
	}
	return nil
}
