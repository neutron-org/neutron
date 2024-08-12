package dex_state_test

import (
	"fmt"
	"testing"
)

type existingUsers int

const (
	none existingUsers = iota
	creator
	oneOther
)

type liquidityDistribution int

const (
	TokenA1TokenB0 liquidityDistribution = iota
	TokenA1TokenB1
	TokenA0TokenB0
	TokenA0TokenB1
)

func setupDepositState(state depositState) {

}

type testStates struct {
	key    string
	states []any
}

type depositState struct {
	existingShareHolders  existingUsers
	liquidityDistribution liquidityDistribution
}

var depositTestStates []testStates = []testStates{
	{key: "existingUsers", states: []any{none, creator, oneOther}},
	{key: "liquidityDistribution", states: []any{TokenA1TokenB0,
		TokenA1TokenB1,
		TokenA0TokenB0,
		TokenA0TokenB1}}}

// func generateAllDepositStates(testStates []testStates) []depositState {

// }

type DepositStateParams struct {
}

func generatePermutations(values []testStates) map[string]any {
	var result map[string]any

	// Recursive function to generate permutations
	var generate func(int, []any)
	generate = func(index int, current []any) {
		// Base case: if we've reached the end of the values, add the current combination to the result
		if index == len(values) {
			// Make a copy of the current slice and add it to the result
			temp := make([]any, len(current))
			copy(temp, current)
			result[values[index].key] = result
			return
		}

		// Iterate over the elements in the current sub-array
		for _, value := range values[index].states {
			// Add the current value to the combination and recurse
			generate(index+1, append(current, value))
		}
	}

	// Start the recursion with an empty combination
	generate(0, []any{})

	return result
}

func TestShit(t *testing.T) {

	permutations := generatePermutations(depositTestStates)

	// Print the permutations
	for i, p := range permutations {
		fmt.Printf("%v: %v\n", i, p)
	}
	t.Fail()
}

// 1. State Conditions
// 1. Existing pool share holders
// 1. None
// 2. Creator
// 3. 1 pre-existing
// 2. Pool liquidity distribution
// 1. 1x tokenA 0x TokenB
// 2. 1x TokenA 1x TokenB
// 3. 0x TokenA 0x TokenB
// 4. 0x TokenA 1x TokenB
// 2. Assertions
// 1. Correct # of shares issued
