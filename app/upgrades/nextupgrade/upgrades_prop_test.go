package nextupgrade

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	appparams "github.com/neutron-org/neutron/v10/app/params"
)

// This file contains property-based tests for calcRedelegations.
//
// Unlike example-based unit tests, these tests generate many random inputs
// (unique delegations, unique target validator sets, and arbitrary set overlap)
// and verify high-level invariants such as:
// - no zero-amount redelegations
// - no self-redelegations
// - stake conservation (total in == total out)
// - target validators end up with (approximately) equal final stake
//
// These tests are meant to catch corner cases that are hard to cover with a
// small number of hand-crafted scenarios.

// redelegationInput is a random input for property-based tests of calcRedelegations.
type redelegationInput struct {
	Delegations   []stakingtypes.DelegationResponse
	NewValidators []sdk.ValAddress
}

// Generate implements testing/quick.Generator.
func (redelegationInput) Generate(r *rand.Rand, _ int) reflect.Value {
	// Number of delegations: [15, 25].
	nDeleg := 15 + r.Intn(10)

	// Number of new validators: [5, 8].
	nNew := 5 + r.Intn(3)

	// Helper to generate unique validator addresses.
	used := make(map[string]struct{})
	nextVal := func() sdk.ValAddress {
		for {
			b := byte(r.Intn(250) + 1)
			v := makeValAddr(b)
			if _, ok := used[v.String()]; !ok {
				used[v.String()] = struct{}{}
				return v
			}
		}
	}

	// Old (delegating) validators: all unique.
	delegVals := make([]sdk.ValAddress, 0, nDeleg)
	for i := 0; i < nDeleg; i++ {
		delegVals = append(delegVals, nextVal())
	}

	// New validators: all unique, with random intersection with delegVals.
	newVals := make([]sdk.ValAddress, 0, nNew)
	for i := 0; i < nNew; i++ {
		if r.Float64() < 0.6 && len(delegVals) > 0 {
			// With some probability reuse a delegating validator
			// to ensure intersections, including potential full subset.
			v := delegVals[r.Intn(len(delegVals))]
			if _, ok := used[v.String()]; !ok {
				used[v.String()] = struct{}{}
				newVals = append(newVals, v)
				continue
			}
		}
		newVals = append(newVals, nextVal())
	}

	// Delegations: positive integer shares in [1, 1000].
	delegs := make([]stakingtypes.DelegationResponse, 0, nDeleg)
	for _, v := range delegVals {
		amount := 1 + r.Intn(1000)
		delegs = append(delegs, makeDelegation(v, fmt.Sprintf("%d", amount)))
	}

	out := redelegationInput{
		Delegations:   delegs,
		NewValidators: newVals,
	}
	return reflect.ValueOf(out)
}

// finalStakePerNewValidator computes the final stake for each new validator as:
// initial delegation (if any) + sum of incoming redelegation messages.
func finalStakePerNewValidator(
	delegations []stakingtypes.DelegationResponse,
	newVals []sdk.ValAddress,
	msgs []stakingtypes.MsgBeginRedelegate,
) map[string]math.Int {
	initial := make(map[string]math.Int)
	for _, d := range delegations {
		initialStake := initial[d.Delegation.ValidatorAddress]
		if initialStake.IsNil() {
			initialStake = math.ZeroInt()
		}
		initial[d.Delegation.ValidatorAddress] = initialStake.Add(d.Balance.Amount)
	}

	final := make(map[string]math.Int)
	for _, v := range newVals {
		initialStake := initial[v.String()]
		if initialStake.IsNil() {
			initialStake = math.ZeroInt()
		}
		final[v.String()] = initialStake
	}

	for _, msg := range msgs {
		finalStake := final[msg.ValidatorDstAddress]
		if finalStake.IsNil() {
			finalStake = math.ZeroInt()
		}
		final[msg.ValidatorDstAddress] = finalStake.Add(
			msg.Amount.Amount,
		)
	}

	return final
}

// maxMinDiff returns max(value) - min(value) over the map, or 0 if empty.
func maxMinDiff(m map[string]math.Int) math.Int {
	first := true
	var min, max math.Int
	for _, v := range m {
		if first {
			min, max = v, v
			first = false
			continue
		}
		if v.LT(min) {
			min = v
		}
		if v.GT(max) {
			max = v
		}
	}
	if first {
		return math.ZeroInt()
	}
	return max.Sub(min)
}

func TestCalcRedelegations_PropertyBased(t *testing.T) {
	prop := func(in redelegationInput) bool {
		// Run calcRedelegations and flatten messages.
		msgs := calcRedelegations(in.Delegations, in.NewValidators, appparams.DefaultDenom)

		// Build lookup sets for old and new validators.
		oldSet := make(map[string]struct{})
		for _, d := range in.Delegations {
			oldSet[d.Delegation.ValidatorAddress] = struct{}{}
		}
		newSet := make(map[string]struct{})
		for _, v := range in.NewValidators {
			newSet[v.String()] = struct{}{}
		}

		// Basic invariants on each message:
		for _, msg := range msgs {
			// 1) no zero-amount redelegation messages
			if msg.Amount.Amount.IsZero() {
				t.Logf("zero-amount msg: %+v", msg)
				return false
			}
			// 2) no self-redelegations
			if msg.ValidatorSrcAddress == msg.ValidatorDstAddress {
				t.Logf("self-redelegation msg: %+v", msg)
				return false
			}
			// 3) src must always belong to the original delegation set
			if _, ok := oldSet[msg.ValidatorSrcAddress]; !ok {
				t.Logf("src %s not in old set", msg.ValidatorSrcAddress)
				return false
			}
			// 4) dst must always belong to the new validator set
			if _, ok := newSet[msg.ValidatorDstAddress]; !ok {
				t.Logf("dst %s not in new set", msg.ValidatorDstAddress)
				return false
			}
		}

		// 1) final stake of each new validator is approximately equal
		final := finalStakePerNewValidator(in.Delegations, in.NewValidators, msgs)
		diff := maxMinDiff(final)
		// Allow a small rounding-related difference (configured by the checks below).
		if diff.GTE(math.NewInt(10)) {
			t.Logf("final stake diff too large: %s", diff)
			return false
		}

		// 2) if a new validator was already present in delegations and had an initial stake
		//    smaller than its final stake, then it must not have any outgoing redelegations.
		initial := make(map[string]math.Int)
		for _, d := range in.Delegations {
			initialStake := initial[d.Delegation.ValidatorAddress]
			if initialStake.IsNil() {
				initialStake = math.ZeroInt()
			}
			initial[d.Delegation.ValidatorAddress] = initialStake.Add(d.Balance.Amount)
		}

		outBySrc := make(map[string]math.Int)
		for _, msg := range msgs {
			out := outBySrc[msg.ValidatorSrcAddress]
			if out.IsNil() {
				out = math.ZeroInt()
			}
			outBySrc[msg.ValidatorSrcAddress] = out.Add(msg.Amount.Amount)
		}

		for _, v := range in.NewValidators {
			vStr := v.String()
			initStake := initial[vStr] // may be 0
			if initStake.IsNil() {
				initStake = math.ZeroInt()
			}
			finStake := final[vStr] // final stake
			if finStake.IsNil() {
				finStake = math.ZeroInt()
			}
			out := outBySrc[vStr] // outgoing amount
			if out.IsNil() {
				out = math.ZeroInt()
			}
			if finStake.GT(initStake) && !out.IsZero() {
				t.Logf("validator %s had initial stake %s < final %s but has outflow %s",
					vStr, initStake, finStake, out)
				return false
			}
		}

		// Additional invariant: total stake is conserved (total in == total out).
		totalInitial := math.ZeroInt()
		for _, d := range in.Delegations {
			totalInitial = totalInitial.Add(d.Balance.Amount)
		}
		totalFinal := math.ZeroInt()
		for _, v := range in.NewValidators {
			totalFinal = totalFinal.Add(final[v.String()])
		}
		if !totalInitial.Equal(totalFinal) {
			t.Logf("total stake mismatch: initial=%s final=%s", totalInitial, totalFinal)
			return false
		}

		// Additional invariant: no duplicate (src,dst) pairs
		seen := make(map[string]struct{})
		for _, msg := range msgs {
			key := msg.ValidatorSrcAddress + "->" + msg.ValidatorDstAddress
			if _, ok := seen[key]; ok {
				t.Logf("duplicate (src,dst) pair: %s (msg=%+v)", key, msg)
				return false
			}
			seen[key] = struct{}{}
		}

		return true
	}

	cfg := &quick.Config{
		MaxCount: 2000,
	}

	if err := quick.Check(prop, cfg); err != nil {
		t.Fatalf("property failed: %v", err)
	}
}
