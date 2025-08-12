package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	math_utils "github.com/neutron-org/neutron/v8/utils/math"
	dextypes "github.com/neutron-org/neutron/v8/x/dex/types"
)

func TestCalcGreatestMatchingRatioBothReservesNonZero(t *testing.T) {
	trueAmount0, trueAmount1 := dextypes.CalcGreatestMatchingRatio(
		math_utils.NewPrecDec(10),
		math_utils.NewPrecDec(40),
		math_utils.NewPrecDec(100),
		math_utils.NewPrecDec(100),
	)
	assert.Equal(t, math_utils.NewPrecDec(25), trueAmount0)
	assert.Equal(t, math_utils.NewPrecDec(100), trueAmount1)
}

func TestCalcGreatestMatchingRatioBothReservesZero(t *testing.T) {
	trueAmount0, trueAmount1 := dextypes.CalcGreatestMatchingRatio(
		math_utils.NewPrecDec(0),
		math_utils.NewPrecDec(0),
		math_utils.NewPrecDec(100),
		math_utils.NewPrecDec(100),
	)
	assert.Equal(t, math_utils.NewPrecDec(100), trueAmount0)
	assert.Equal(t, math_utils.NewPrecDec(100), trueAmount1)
}

func TestCalcGreatestMatchingRatioWrongCoinDeposited(t *testing.T) {
	trueAmount0, trueAmount1 := dextypes.CalcGreatestMatchingRatio(
		math_utils.NewPrecDec(100),
		math_utils.NewPrecDec(0),
		math_utils.NewPrecDec(0),
		math_utils.NewPrecDec(100),
	)
	assert.Equal(t, math_utils.ZeroPrecDec(), trueAmount0)
	assert.Equal(t, math_utils.ZeroPrecDec(), trueAmount1)

	trueAmount0, trueAmount1 = dextypes.CalcGreatestMatchingRatio(
		math_utils.NewPrecDec(0),
		math_utils.NewPrecDec(100),
		math_utils.NewPrecDec(100),
		math_utils.NewPrecDec(0),
	)
	assert.Equal(t, math_utils.ZeroPrecDec(), trueAmount0)
	assert.Equal(t, math_utils.ZeroPrecDec(), trueAmount1)
}

func TestCalcGreatestMatchingRatioOneReserveZero(t *testing.T) {
	trueAmount0, trueAmount1 := dextypes.CalcGreatestMatchingRatio(
		math_utils.NewPrecDec(100),
		math_utils.NewPrecDec(0),
		math_utils.NewPrecDec(100),
		math_utils.NewPrecDec(100),
	)
	assert.Equal(t, math_utils.NewPrecDec(100), trueAmount0)
	assert.Equal(t, math_utils.ZeroPrecDec(), trueAmount1)

	trueAmount0, trueAmount1 = dextypes.CalcGreatestMatchingRatio(
		math_utils.NewPrecDec(0),
		math_utils.NewPrecDec(100),
		math_utils.NewPrecDec(100),
		math_utils.NewPrecDec(100),
	)
	assert.Equal(t, math_utils.ZeroPrecDec(), trueAmount0)
	assert.Equal(t, math_utils.NewPrecDec(100), trueAmount1)
}

func TestCalcGreatestMatchingRatio2SidedPoolBothSidesRightRatio(t *testing.T) {
	// WHEN deposit into a pool with a ratio of 2:5 with the same ratio all of the tokens are used
	trueAmount0, trueAmount1 := dextypes.CalcGreatestMatchingRatio(
		math_utils.NewPrecDec(20),
		math_utils.NewPrecDec(50),
		math_utils.NewPrecDec(4),
		math_utils.NewPrecDec(10),
	)

	// THEN both amounts are fully user

	assert.Equal(t, math_utils.NewPrecDec(4), trueAmount0)
	assert.Equal(t, math_utils.NewPrecDec(10), trueAmount1)
}

func TestCalcGreatestMatchingRatio2SidedPoolBothSidesWrongRatio(t *testing.T) {
	// WHEN deposit into a pool with a ratio of 3:2 with a ratio of 2:1
	trueAmount0, trueAmount1 := dextypes.CalcGreatestMatchingRatio(
		math_utils.NewPrecDec(3),
		math_utils.NewPrecDec(2),
		math_utils.NewPrecDec(20),
		math_utils.NewPrecDec(10),
	)

	// THEN all of Token1 is used and 3/4 of token0 is used

	assert.Equal(t, math_utils.NewPrecDec(15), trueAmount0)
	assert.Equal(t, math_utils.NewPrecDec(10), trueAmount1)
}

func TestCalcGreatestMatchingRatio2SidedPoolBothSidesWrongRatio2(t *testing.T) {
	// IF deposit into a pool with a ratio of 2:3 with a ratio of 1:2
	trueAmount0, trueAmount1 := dextypes.CalcGreatestMatchingRatio(
		math_utils.NewPrecDec(2),
		math_utils.NewPrecDec(3),
		math_utils.NewPrecDec(10),
		math_utils.NewPrecDec(20),
	)

	// THEN all of Token0 is used and 3/4 of token1 is used

	assert.Equal(t, math_utils.NewPrecDec(10), trueAmount0)
	assert.Equal(t, math_utils.NewPrecDec(15), trueAmount1)
}

func TestCalcGreatestMatchingRatio1SidedPoolBothSides(t *testing.T) {
	// WHEN deposit Token0 and Token1 into a pool with only Token0
	trueAmount0, trueAmount1 := dextypes.CalcGreatestMatchingRatio(
		math_utils.NewPrecDec(10),
		math_utils.NewPrecDec(0),
		math_utils.NewPrecDec(10),
		math_utils.NewPrecDec(10),
	)

	// THEN only Token0 is used

	assert.Equal(t, math_utils.NewPrecDec(10), trueAmount0)
	assert.Equal(t, math_utils.ZeroPrecDec(), trueAmount1)
}

func TestCalcGreatestMatchingRatio1SidedPoolBothSides2(t *testing.T) {
	// WHEN deposit Token0 and Token1 into a pool with only Token1
	trueAmount0, trueAmount1 := dextypes.CalcGreatestMatchingRatio(

		math_utils.NewPrecDec(0),
		math_utils.NewPrecDec(10),
		math_utils.NewPrecDec(10),
		math_utils.NewPrecDec(10),
	)

	// THEN only Token1 is used

	assert.Equal(t, math_utils.ZeroPrecDec(), trueAmount0)
	assert.Equal(t, math_utils.NewPrecDec(10), trueAmount1)
}

func TestCalcGreatestMatchingRatio1SidedPool1SidedToken0(t *testing.T) {
	// WHEN deposit Token0 into a pool with only Token1
	trueAmount0, trueAmount1 := dextypes.CalcGreatestMatchingRatio(

		math_utils.NewPrecDec(0),
		math_utils.NewPrecDec(10),
		math_utils.NewPrecDec(10),
		math_utils.NewPrecDec(0),
	)

	// THEN no amounts are used

	assert.Equal(t, math_utils.ZeroPrecDec(), trueAmount0)
	assert.Equal(t, math_utils.ZeroPrecDec(), trueAmount1)
}

func TestCalcGreatestMatchingRatio1SidedPool1SidedToken0B(t *testing.T) {
	// WHEN deposit Token0 into a pool with only Token0
	trueAmount0, trueAmount1 := dextypes.CalcGreatestMatchingRatio(
		math_utils.NewPrecDec(10),
		math_utils.NewPrecDec(0),
		math_utils.NewPrecDec(10),
		math_utils.NewPrecDec(0),
	)

	// THEN all of Token0 is used

	assert.Equal(t, math_utils.NewPrecDec(10), trueAmount0)
	assert.Equal(t, math_utils.ZeroPrecDec(), trueAmount1)
}

func TestCalcGreatestMatchingRatio1SidedPool1SidedToken1(t *testing.T) {
	// WHEN deposit Token1 into a pool with only Token0
	trueAmount0, trueAmount1 := dextypes.CalcGreatestMatchingRatio(
		math_utils.NewPrecDec(10),
		math_utils.NewPrecDec(0),
		math_utils.NewPrecDec(0),
		math_utils.NewPrecDec(1),
	)

	// THEN no amounts are used

	assert.Equal(t, math_utils.ZeroPrecDec(), trueAmount0)
	assert.Equal(t, math_utils.ZeroPrecDec(), trueAmount1)
}

func TestCalcGreatestMatchingRatio1SidedPool1SidedToken1B(t *testing.T) {
	// WHEN deposit Token1 into a pool with only Token1
	trueAmount0, trueAmount1 := dextypes.CalcGreatestMatchingRatio(
		math_utils.NewPrecDec(0),
		math_utils.NewPrecDec(10),
		math_utils.NewPrecDec(0),
		math_utils.NewPrecDec(10),
	)

	// THEN all of Token1 is used

	assert.Equal(t, math_utils.ZeroPrecDec(), trueAmount0)
	assert.Equal(t, math_utils.NewPrecDec(10), trueAmount1)
}
