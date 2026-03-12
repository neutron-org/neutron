package nextupgrade

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	appparams "github.com/neutron-org/neutron/v10/app/params"
	"github.com/stretchr/testify/require"
)

var (
	testDelegator = sdk.AccAddress(make([]byte, 20)).String()
)

// makeValAddr creates a deterministic sdk.ValAddress from a small integer, useful for tests.
func makeValAddr(n byte) sdk.ValAddress {
	addr := make([]byte, 20)
	addr[0] = n
	return sdk.ValAddress(addr)
}

// makeDelegation creates a Delegation with the given validator address and share amount.
func makeDelegation(val sdk.ValAddress, shares string) stakingtypes.Delegation {
	return stakingtypes.Delegation{
		DelegatorAddress: testDelegator,
		ValidatorAddress: val.String(),
		Shares:           math.LegacyMustNewDecFromStr(shares),
	}
}

// calcRedelegationsHelper wraps calcRedelegations with the test denom constants.
func calcRedelegationsHelper(delegations []stakingtypes.Delegation, newValidators []sdk.ValAddress) []Redelegation {
	return calcRedelegations(delegations, newValidators, appparams.DefaultDenom)
}

// totalRedelegated returns the sum of Redelegated across all Redelegation entries.
func totalRedelegated(rs []Redelegation) math.LegacyDec {
	sum := math.LegacyZeroDec()
	for _, r := range rs {
		sum = sum.Add(r.Redelegated)
	}
	return sum
}

// srcSharesFor returns the sum of all MsgBeginRedelegate amounts originating from a given
// source validator across the entire redelegation plan.
func srcSharesFor(rs []Redelegation, srcVal sdk.ValAddress) math.Int {
	sum := math.ZeroInt()
	for _, r := range rs {
		for _, msg := range r.RedelegationMsgs {
			if msg.ValidatorSrcAddress == srcVal.String() {
				sum = sum.Add(msg.Amount.Amount)
			}
		}
	}
	return sum
}

func TestCalcRedelegations_NoDelegations(t *testing.T) {
	newVals := []sdk.ValAddress{makeValAddr(10), makeValAddr(11)}
	result := calcRedelegationsHelper(nil, newVals)
	require.Len(t, result, 2)
	for _, r := range result {
		require.True(t, r.Redelegated.IsZero())
		require.Empty(t, r.RedelegationMsgs)
	}
}

// One old validator → one new validator: entire stake moves in a single message.
func TestCalcRedelegations_OneToOne(t *testing.T) {
	oldVal := makeValAddr(1)
	newVal := makeValAddr(10)

	delegations := []stakingtypes.Delegation{
		makeDelegation(oldVal, "100"),
	}
	result := calcRedelegationsHelper(delegations, []sdk.ValAddress{newVal})

	require.Len(t, result, 1)
	require.Equal(t, newVal, result[0].ValidatorAddress)
	require.Len(t, result[0].RedelegationMsgs, 1)

	msg := result[0].RedelegationMsgs[0]
	require.Equal(t, testDelegator, msg.DelegatorAddress)
	require.Equal(t, oldVal.String(), msg.ValidatorSrcAddress)
	require.Equal(t, newVal.String(), msg.ValidatorDstAddress)
	require.Equal(t, math.NewInt(100), msg.Amount.Amount)
	require.Equal(t, appparams.DefaultDenom, msg.Amount.Denom)
	require.Equal(t, math.LegacyNewDec(100), result[0].Redelegated)
}

// One large old delegation is split evenly across several new validators.
func TestCalcRedelegations_OneOldManyNew_EvenSplit(t *testing.T) {
	oldVal := makeValAddr(1)
	newVals := []sdk.ValAddress{makeValAddr(10), makeValAddr(11), makeValAddr(12)}
	expectedAmountPer := math.LegacyNewDec(100) // total = 300

	delegations := []stakingtypes.Delegation{
		makeDelegation(oldVal, "300"),
	}
	result := calcRedelegationsHelper(delegations, newVals)

	require.Len(t, result, 3)
	for i, r := range result {
		require.Equal(t, newVals[i], r.ValidatorAddress, "validator index %d", i)
		require.Len(t, r.RedelegationMsgs, 1)

		msg := r.RedelegationMsgs[0]
		require.Equal(t, testDelegator, msg.DelegatorAddress)
		require.Equal(t, oldVal.String(), msg.ValidatorSrcAddress)
		require.Equal(t, newVals[i].String(), msg.ValidatorDstAddress)
		require.Equal(t, expectedAmountPer.TruncateInt(), msg.Amount.Amount)
	}
}

// Several old validators (each holding exactly the per-validator share) map one-to-one to
// new validators.
func TestCalcRedelegations_ManyOldManyNew_OneToOneMapping(t *testing.T) {
	oldVal1, oldVal2 := makeValAddr(1), makeValAddr(2)
	newVal1, newVal2 := makeValAddr(10), makeValAddr(11)
	expectedAmountPer := math.LegacyNewDec(50)

	delegations := []stakingtypes.Delegation{
		makeDelegation(oldVal1, "50"),
		makeDelegation(oldVal2, "50"),
	}
	result := calcRedelegationsHelper(delegations, []sdk.ValAddress{newVal1, newVal2})

	require.Len(t, result, 2)

	msg0 := result[0].RedelegationMsgs[0]
	require.Equal(t, oldVal1.String(), msg0.ValidatorSrcAddress)
	require.Equal(t, newVal1.String(), msg0.ValidatorDstAddress)
	require.Equal(t, expectedAmountPer.TruncateInt(), msg0.Amount.Amount)

	msg1 := result[1].RedelegationMsgs[0]
	require.Equal(t, oldVal2.String(), msg1.ValidatorSrcAddress)
	require.Equal(t, newVal2.String(), msg1.ValidatorDstAddress)
	require.Equal(t, expectedAmountPer.TruncateInt(), msg1.Amount.Amount)
}

// Multiple old validators with varying amounts; delegations need to be merged and split.
func TestCalcRedelegations_MixedSplitAndMerge(t *testing.T) {
	// total = 100, 2 new validators → 50 each
	oldVal1 := makeValAddr(1) // 30
	oldVal2 := makeValAddr(2) // 40
	oldVal3 := makeValAddr(3) // 30

	newVal1 := makeValAddr(10)
	newVal2 := makeValAddr(11)
	expectedAmountPer := math.LegacyNewDec(50)

	delegations := []stakingtypes.Delegation{
		makeDelegation(oldVal1, "30"),
		makeDelegation(oldVal2, "40"),
		makeDelegation(oldVal3, "30"),
	}
	result := calcRedelegationsHelper(delegations, []sdk.ValAddress{newVal1, newVal2})

	require.Len(t, result, 2)
	require.Equal(t, math.LegacyNewDec(100), totalRedelegated(result))

	// newVal1 receives 30 from oldVal1 and 20 from oldVal2.
	require.Equal(t, newVal1, result[0].ValidatorAddress)
	require.Len(t, result[0].RedelegationMsgs, 2)
	require.Equal(t, oldVal1.String(), result[0].RedelegationMsgs[0].ValidatorSrcAddress)
	require.Equal(t, newVal1.String(), result[0].RedelegationMsgs[0].ValidatorDstAddress)
	require.Equal(t, math.NewInt(30), result[0].RedelegationMsgs[0].Amount.Amount)
	require.Equal(t, oldVal2.String(), result[0].RedelegationMsgs[1].ValidatorSrcAddress)
	require.Equal(t, newVal1.String(), result[0].RedelegationMsgs[1].ValidatorDstAddress)
	require.Equal(t, math.NewInt(20), result[0].RedelegationMsgs[1].Amount.Amount)
	require.Equal(t, expectedAmountPer.TruncateInt(), result[0].Redelegated.TruncateInt())

	// newVal2 receives 20 from oldVal2 and 30 from oldVal3.
	require.Equal(t, newVal2, result[1].ValidatorAddress)
	require.Len(t, result[1].RedelegationMsgs, 2)
	require.Equal(t, oldVal2.String(), result[1].RedelegationMsgs[0].ValidatorSrcAddress)
	require.Equal(t, newVal2.String(), result[1].RedelegationMsgs[0].ValidatorDstAddress)
	require.Equal(t, math.NewInt(20), result[1].RedelegationMsgs[0].Amount.Amount)
	require.Equal(t, oldVal3.String(), result[1].RedelegationMsgs[1].ValidatorSrcAddress)
	require.Equal(t, newVal2.String(), result[1].RedelegationMsgs[1].ValidatorDstAddress)
	require.Equal(t, math.NewInt(30), result[1].RedelegationMsgs[1].Amount.Amount)
	require.Equal(t, expectedAmountPer.TruncateInt(), result[1].Redelegated.TruncateInt())
}

// Total shares do not divide evenly: the last new validator absorbs the remainder.
func TestCalcRedelegations_UnevenTotal_LastValidatorAbsorbsRemainder(t *testing.T) {
	oldVal := makeValAddr(1)
	newVals := []sdk.ValAddress{makeValAddr(10), makeValAddr(11), makeValAddr(12)}
	// 100 / 3 = 33.333...
	expectedAmountPer := math.LegacyNewDec(100).Quo(math.LegacyNewDec(3))

	delegations := []stakingtypes.Delegation{
		makeDelegation(oldVal, "100"),
	}
	result := calcRedelegationsHelper(delegations, newVals)

	require.Len(t, result, 3)

	require.Equal(t, expectedAmountPer, result[0].Redelegated)
	require.Equal(t, expectedAmountPer, result[1].Redelegated)

	// Last validator gets whatever remains after the first two took their exact share.
	expectedLast := math.LegacyNewDec(100).Sub(expectedAmountPer).Sub(expectedAmountPer)
	require.Equal(t, expectedLast, result[2].Redelegated)

	// Conservation: all 100 shares must be accounted for.
	require.Equal(t, math.LegacyNewDec(100), totalRedelegated(result))
}

// Source-validator amounts are fully consumed across a larger mixed scenario.
func TestCalcRedelegations_SourceAmountsAreFullyConsumed(t *testing.T) {
	oldVal1 := makeValAddr(1) // 70
	oldVal2 := makeValAddr(2) // 80
	oldVal3 := makeValAddr(3) // 50
	// total = 200, 4 new validators → 50 each (exact)
	newVals := []sdk.ValAddress{makeValAddr(10), makeValAddr(11), makeValAddr(12), makeValAddr(13)}
	expectedAmountPer := math.LegacyNewDec(50)

	delegations := []stakingtypes.Delegation{
		makeDelegation(oldVal1, "70"),
		makeDelegation(oldVal2, "80"),
		makeDelegation(oldVal3, "50"),
	}
	result := calcRedelegationsHelper(delegations, newVals)

	for i, r := range result {
		require.Equal(t, expectedAmountPer, r.Redelegated, "validator index %d", i)
	}

	require.Equal(t, math.NewInt(70), srcSharesFor(result, oldVal1))
	require.Equal(t, math.NewInt(80), srcSharesFor(result, oldVal2))
	require.Equal(t, math.NewInt(50), srcSharesFor(result, oldVal3))
}

// Five new validators (matches NewMaxValidators), 18-validator old set with varying amounts.
func TestCalcRedelegations_FullValidatorSetRedistribution(t *testing.T) {
	type oldValSpec struct {
		addr   sdk.ValAddress
		shares string
	}
	specs := []oldValSpec{
		{makeValAddr(1), "10"},
		{makeValAddr(2), "15"},
		{makeValAddr(3), "8"},
		{makeValAddr(4), "25"},
		{makeValAddr(5), "30"},
		{makeValAddr(6), "12"},
		{makeValAddr(7), "45"},
		{makeValAddr(8), "60"},
		{makeValAddr(9), "20"},
		{makeValAddr(10), "35"},
		{makeValAddr(11), "18"},
		{makeValAddr(12), "50"},
		{makeValAddr(13), "28"},
		{makeValAddr(14), "40"},
		{makeValAddr(15), "55"},
		{makeValAddr(16), "22"},
		{makeValAddr(17), "62"},
		{makeValAddr(18), "65"},
	}
	newVals := []sdk.ValAddress{
		makeValAddr(20), makeValAddr(21), makeValAddr(22),
		makeValAddr(23), makeValAddr(24),
	}
	// total = 600, 5 new validators → 120 each (exact)
	expectedAmountPer := math.LegacyNewDec(120)

	delegations := make([]stakingtypes.Delegation, len(specs))
	for i, s := range specs {
		delegations[i] = makeDelegation(s.addr, s.shares)
	}

	result := calcRedelegationsHelper(delegations, newVals)

	require.Len(t, result, 5)
	for i, r := range result {
		require.Equal(t, newVals[i], r.ValidatorAddress)
		require.Equal(t, expectedAmountPer, r.Redelegated, "new validator index %d should receive exactly %s shares", i, expectedAmountPer)
	}

	require.Equal(t, math.LegacyNewDec(600), totalRedelegated(result))

	// Every old validator's shares must be fully redistributed.
	for _, s := range specs {
		expected, _ := math.NewIntFromString(s.shares)
		require.Equal(t, expected, srcSharesFor(result, s.addr), "old validator %s shares not fully consumed", s.addr)
	}
}

// Every message must carry the correct delegator address and denom.
func TestCalcRedelegations_MsgFieldsAreFullyPopulated(t *testing.T) {
	oldVal := makeValAddr(1)
	newVals := []sdk.ValAddress{makeValAddr(10), makeValAddr(11)}

	delegations := []stakingtypes.Delegation{
		makeDelegation(oldVal, "100"),
	}
	result := calcRedelegationsHelper(delegations, newVals)

	for _, r := range result {
		for _, msg := range r.RedelegationMsgs {
			require.Equal(t, testDelegator, msg.DelegatorAddress)
			require.Equal(t, appparams.DefaultDenom, msg.Amount.Denom)
			require.False(t, msg.Amount.Amount.IsZero())
		}
	}
}

// Destination in each MsgBeginRedelegate must match the enclosing Redelegation.ValidatorAddress.
func TestCalcRedelegations_DstAddressMatchesRedelegationValidator(t *testing.T) {
	oldVals := []sdk.ValAddress{makeValAddr(1), makeValAddr(2)}
	newVals := []sdk.ValAddress{makeValAddr(10), makeValAddr(11)}

	delegations := []stakingtypes.Delegation{
		makeDelegation(oldVals[0], "50"),
		makeDelegation(oldVals[1], "50"),
	}
	result := calcRedelegationsHelper(delegations, newVals)

	for _, r := range result {
		for _, msg := range r.RedelegationMsgs {
			require.Equal(t, r.ValidatorAddress.String(), msg.ValidatorDstAddress,
				"ValidatorDstAddress must equal the enclosing Redelegation.ValidatorAddress")
		}
	}
}

// Source in each MsgBeginRedelegate must always be an old (input) validator.
func TestCalcRedelegations_SrcAddressIsAlwaysOldValidator(t *testing.T) {
	oldVals := []sdk.ValAddress{makeValAddr(1), makeValAddr(2)}
	newVals := []sdk.ValAddress{makeValAddr(10), makeValAddr(11)}

	oldValSet := map[string]struct{}{
		oldVals[0].String(): {},
		oldVals[1].String(): {},
	}

	delegations := []stakingtypes.Delegation{
		makeDelegation(oldVals[0], "50"),
		makeDelegation(oldVals[1], "50"),
	}
	result := calcRedelegationsHelper(delegations, newVals)

	for _, r := range result {
		for _, msg := range r.RedelegationMsgs {
			_, isOld := oldValSet[msg.ValidatorSrcAddress]
			require.True(t, isOld, "ValidatorSrcAddress %s is not an old validator", msg.ValidatorSrcAddress)
		}
	}
}
