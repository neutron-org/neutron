package keeper

import (
	"cosmossdk.io/math"
)

func SortAmounts(tokenA, token0 string, amountsA, amountsB []math.Int) ([]math.Int, []math.Int) {
	if tokenA == token0 {
		return amountsA, amountsB
	}

	return amountsB, amountsA
}

func GetInOutTokens(tokenIn, tokenA, tokenB string) (_, tokenOut string) {
	if tokenIn == tokenA {
		return tokenA, tokenB
	}

	return tokenB, tokenA
}

func NormalizeAllTickIndexes(takerDenom, token0 string, tickIndexes []int64) []int64 {
	if takerDenom != token0 {
		result := make([]int64, len(tickIndexes))
		for i, idx := range tickIndexes {
			result[i] = idx * -1
		}
		return result
	}

	// NB: does not return a different slice because no change
	return tickIndexes
}
