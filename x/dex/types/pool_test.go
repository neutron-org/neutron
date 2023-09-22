package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/neutron-org/neutron/x/dex/types"
	"github.com/stretchr/testify/assert"
)

func TestCalcGreatestMatchingRatioBothReservesNonZero(t *testing.T) {
	trueAmount0, trueAmount1 := CalcGreatestMatchingRatio(
		sdk.NewInt(10),
		sdk.NewInt(40),
		sdk.NewInt(100),
		sdk.NewInt(100),
	)
	assert.Equal(t, sdk.NewInt(25), trueAmount0)
	assert.Equal(t, sdk.NewInt(100), trueAmount1)
}

func TestCalcGreatestMatchingRatioBothReservesZero(t *testing.T) {
	trueAmount0, trueAmount1 := CalcGreatestMatchingRatio(
		sdk.NewInt(0),
		sdk.NewInt(0),
		sdk.NewInt(100),
		sdk.NewInt(100),
	)
	assert.Equal(t, sdk.NewInt(100), trueAmount0)
	assert.Equal(t, sdk.NewInt(100), trueAmount1)
}

func TestCalcGreatestMatchingRatioWrongCoinDeposited(t *testing.T) {
	trueAmount0, trueAmount1 := CalcGreatestMatchingRatio(
		sdk.NewInt(100),
		sdk.NewInt(0),
		sdk.NewInt(0),
		sdk.NewInt(100),
	)
	assert.Equal(t, sdk.NewInt(0), trueAmount0)
	assert.Equal(t, sdk.NewInt(0), trueAmount1)

	trueAmount0, trueAmount1 = CalcGreatestMatchingRatio(
		sdk.NewInt(0),
		sdk.NewInt(100),
		sdk.NewInt(100),
		sdk.NewInt(0),
	)
	assert.Equal(t, sdk.NewInt(0), trueAmount0)
	assert.Equal(t, sdk.NewInt(0), trueAmount1)
}

func TestCalcGreatestMatchingRatioOneReserveZero(t *testing.T) {
	trueAmount0, trueAmount1 := CalcGreatestMatchingRatio(
		sdk.NewInt(100),
		sdk.NewInt(0),
		sdk.NewInt(100),
		sdk.NewInt(100),
	)
	assert.Equal(t, sdk.NewInt(100), trueAmount0)
	assert.Equal(t, sdk.NewInt(0), trueAmount1)

	trueAmount0, trueAmount1 = CalcGreatestMatchingRatio(
		sdk.NewInt(0),
		sdk.NewInt(100),
		sdk.NewInt(100),
		sdk.NewInt(100),
	)
	assert.Equal(t, sdk.NewInt(0), trueAmount0)
	assert.Equal(t, sdk.NewInt(100), trueAmount1)
}

func TestCalcGreatestMatchingRatio2SidedPoolBothSidesRightRatio(t *testing.T) {
	// WHEN deposit into a pool with a ratio of 2:5 with the same ratio all of the tokens are used
	trueAmount0, trueAmount1 := CalcGreatestMatchingRatio(
		sdk.NewInt(20),
		sdk.NewInt(50),
		sdk.NewInt(4),
		sdk.NewInt(10),
	)

	// THEN both amounts are fully user

	assert.Equal(t, sdk.NewInt(4), trueAmount0)
	assert.Equal(t, sdk.NewInt(10), trueAmount1)
}

func TestCalcGreatestMatchingRatio2SidedPoolBothSidesWrongRatio(t *testing.T) {
	// WHEN deposit into a pool with a ratio of 3:2 with a ratio of 2:1
	trueAmount0, trueAmount1 := CalcGreatestMatchingRatio(
		sdk.NewInt(3),
		sdk.NewInt(2),
		sdk.NewInt(20),
		sdk.NewInt(10),
	)

	// THEN all of Token1 is used and 3/4 of token0 is used

	assert.Equal(t, sdk.NewInt(15), trueAmount0)
	assert.Equal(t, sdk.NewInt(10), trueAmount1)
}

func TestCalcGreatestMatchingRatio2SidedPoolBothSidesWrongRatio2(t *testing.T) {
	// IF deposit into a pool with a ratio of 2:3 with a ratio of 1:2
	trueAmount0, trueAmount1 := CalcGreatestMatchingRatio(
		sdk.NewInt(2),
		sdk.NewInt(3),
		sdk.NewInt(10),
		sdk.NewInt(20),
	)

	// THEN all of Token0 is used and 3/4 of token1 is used

	assert.Equal(t, sdk.NewInt(10), trueAmount0)
	assert.Equal(t, sdk.NewInt(15), trueAmount1)
}

func TestCalcGreatestMatchingRatio1SidedPoolBothSides(t *testing.T) {
	// WHEN deposit Token0 and Token1 into a pool with only Token0
	trueAmount0, trueAmount1 := CalcGreatestMatchingRatio(
		sdk.NewInt(10),
		sdk.NewInt(0),
		sdk.NewInt(10),
		sdk.NewInt(10),
	)

	// THEN only Token0 is used

	assert.Equal(t, sdk.NewInt(10), trueAmount0)
	assert.Equal(t, sdk.NewInt(0), trueAmount1)
}

func TestCalcGreatestMatchingRatio1SidedPoolBothSides2(t *testing.T) {
	// WHEN deposit Token0 and Token1 into a pool with only Token1
	trueAmount0, trueAmount1 := CalcGreatestMatchingRatio(

		sdk.NewInt(0),
		sdk.NewInt(10),
		sdk.NewInt(10),
		sdk.NewInt(10),
	)

	// THEN only Token1 is used

	assert.Equal(t, sdk.NewInt(0), trueAmount0)
	assert.Equal(t, sdk.NewInt(10), trueAmount1)
}

func TestCalcGreatestMatchingRatio1SidedPool1SidedToken0(t *testing.T) {
	// WHEN deposit Token0 into a pool with only Token1
	trueAmount0, trueAmount1 := CalcGreatestMatchingRatio(

		sdk.NewInt(0),
		sdk.NewInt(10),
		sdk.NewInt(10),
		sdk.NewInt(0),
	)

	// THEN no amounts are used

	assert.Equal(t, sdk.NewInt(0), trueAmount0)
	assert.Equal(t, sdk.NewInt(0), trueAmount1)
}

func TestCalcGreatestMatchingRatio1SidedPool1SidedToken0B(t *testing.T) {
	// WHEN deposit Token0 into a pool with only Token0
	trueAmount0, trueAmount1 := CalcGreatestMatchingRatio(
		sdk.NewInt(10),
		sdk.NewInt(0),
		sdk.NewInt(10),
		sdk.NewInt(0),
	)

	// THEN all of Token0 is used

	assert.Equal(t, sdk.NewInt(10), trueAmount0)
	assert.Equal(t, sdk.NewInt(0), trueAmount1)
}

func TestCalcGreatestMatchingRatio1SidedPool1SidedToken1(t *testing.T) {
	// WHEN deposit Token1 into a pool with only Token0
	trueAmount0, trueAmount1 := CalcGreatestMatchingRatio(
		sdk.NewInt(10),
		sdk.NewInt(0),
		sdk.NewInt(0),
		sdk.NewInt(1),
	)

	// THEN no amounts are used

	assert.Equal(t, sdk.NewInt(0), trueAmount0)
	assert.Equal(t, sdk.NewInt(0), trueAmount1)
}

func TestCalcGreatestMatchingRatio1SidedPool1SidedToken1B(t *testing.T) {
	// WHEN deposit Token1 into a pool with only Token1
	trueAmount0, trueAmount1 := CalcGreatestMatchingRatio(
		sdk.NewInt(0),
		sdk.NewInt(10),
		sdk.NewInt(0),
		sdk.NewInt(10),
	)

	// THEN all of Token1 is used

	assert.Equal(t, sdk.NewInt(0), trueAmount0)
	assert.Equal(t, sdk.NewInt(10), trueAmount1)
}
