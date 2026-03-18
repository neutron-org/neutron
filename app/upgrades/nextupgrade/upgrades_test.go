package nextupgrade

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/neutron-org/neutron/v10/app/params"
)

var testDelegator = sdk.AccAddress(make([]byte, 20)).String()

// makeValAddr creates a deterministic sdk.ValAddress from a small integer, useful for tests.
func makeValAddr(n byte) sdk.ValAddress {
	addr := make([]byte, 20)
	addr[0] = n
	return sdk.ValAddress(addr)
}

// makeDelegation creates a Delegation with the given validator address and share amount.
func makeDelegation(val sdk.ValAddress, shares string) stakingtypes.DelegationResponse {
	return stakingtypes.DelegationResponse{
		Delegation: stakingtypes.Delegation{
			DelegatorAddress: testDelegator,
			ValidatorAddress: val.String(),
			Shares:           math.LegacyMustNewDecFromStr(shares),
		},
		Balance: sdk.NewCoin(appparams.DefaultDenom, math.LegacyMustNewDecFromStr(shares).TruncateInt()),
	}
}

// calcRedelegationsHelper wraps calcRedelegations with the test denom constants.
func calcRedelegationsHelper(delegations []stakingtypes.DelegationResponse, newValidators []sdk.ValAddress) []stakingtypes.MsgBeginRedelegate {
	return calcRedelegations(delegations, newValidators, appparams.DefaultDenom)
}

// totalRedelegated returns the sum of Redelegated across all Redelegation entries.
func totalRedelegated(rs []stakingtypes.MsgBeginRedelegate) math.Int {
	sum := math.NewInt(0)
	for _, r := range rs {
		sum = sum.Add(r.Amount.Amount)
	}
	return sum
}

// srcTokensFor returns the sum of all MsgBeginRedelegate amounts originating from a given
// source validator across the entire redelegation plan.
func srcTokensFor(rs []stakingtypes.MsgBeginRedelegate, srcVal string) math.Int {
	sum := math.ZeroInt()
	for _, r := range rs {
		if r.ValidatorSrcAddress == srcVal {
			sum = sum.Add(r.Amount.Amount)
		}
	}
	return sum
}

// dstTokensFor returns the sum of all MsgBeginRedelegate amounts destined to a given
// destination validator across the entire redelegation plan.
func dstTokensFor(rs []stakingtypes.MsgBeginRedelegate, dstVal string) math.Int {
	sum := math.ZeroInt()
	for _, r := range rs {
		if r.ValidatorDstAddress == dstVal {
			sum = sum.Add(r.Amount.Amount)
		}
	}
	return sum
}

func TestCalcRedelegations_NoDelegations(t *testing.T) {
	newVals := []sdk.ValAddress{makeValAddr(10), makeValAddr(11)}
	result := calcRedelegationsHelper(nil, newVals)
	require.Len(t, result, 0)
}

// One old validator → one new validator: entire stake moves in a single message.
func TestCalcRedelegations_OneToOne(t *testing.T) {
	oldVal := makeValAddr(1)
	newVal := makeValAddr(10)

	delegations := []stakingtypes.DelegationResponse{
		makeDelegation(oldVal, "100"),
	}
	result := calcRedelegationsHelper(delegations, []sdk.ValAddress{newVal})

	require.Len(t, result, 1)

	msg := result[0]
	require.Equal(t, testDelegator, msg.DelegatorAddress)
	require.Equal(t, oldVal.String(), msg.ValidatorSrcAddress)
	require.Equal(t, newVal.String(), msg.ValidatorDstAddress)
	require.Equal(t, math.NewInt(100), msg.Amount.Amount)
	require.Equal(t, appparams.DefaultDenom, msg.Amount.Denom)
	require.Equal(t, math.NewInt(100), msg.Amount.Amount)
}

// One large old delegation is split evenly across several new validators.
func TestCalcRedelegations_OneOldManyNew_EvenSplit(t *testing.T) {
	oldVal := makeValAddr(1)
	newVals := []sdk.ValAddress{makeValAddr(10), makeValAddr(11), makeValAddr(12)}
	expectedAmountPer := math.LegacyNewDec(100) // total = 300

	delegations := []stakingtypes.DelegationResponse{
		makeDelegation(oldVal, "300"),
	}
	result := calcRedelegationsHelper(delegations, newVals)

	require.Len(t, result, 3)
	for i, r := range result {
		require.Equal(t, newVals[i].String(), r.ValidatorDstAddress, "validator index %d", i)

		msg := r
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

	delegations := []stakingtypes.DelegationResponse{
		makeDelegation(oldVal1, "50"),
		makeDelegation(oldVal2, "50"),
	}
	result := calcRedelegationsHelper(delegations, []sdk.ValAddress{newVal1, newVal2})

	require.Len(t, result, 2)

	msg0 := result[0]
	require.Equal(t, oldVal1.String(), msg0.ValidatorSrcAddress)
	require.Equal(t, newVal1.String(), msg0.ValidatorDstAddress)
	require.Equal(t, expectedAmountPer.TruncateInt(), msg0.Amount.Amount)

	msg1 := result[1]
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

	delegations := []stakingtypes.DelegationResponse{
		makeDelegation(oldVal1, "30"),
		makeDelegation(oldVal2, "40"),
		makeDelegation(oldVal3, "30"),
	}
	result := calcRedelegationsHelper(delegations, []sdk.ValAddress{newVal1, newVal2})

	require.Len(t, result, 4)
	require.Equal(t, math.NewInt(100), totalRedelegated(result))

	// newVal1 receives 30 from oldVal1 and 20 from oldVal2.
	require.Equal(t, newVal1.String(), result[0].ValidatorDstAddress)
	require.Equal(t, oldVal1.String(), result[0].ValidatorSrcAddress)
	require.Equal(t, newVal1.String(), result[0].ValidatorDstAddress)
	require.Equal(t, math.NewInt(30), result[0].Amount.Amount)
	require.Equal(t, oldVal2.String(), result[1].ValidatorSrcAddress)
	require.Equal(t, newVal1.String(), result[1].ValidatorDstAddress)
	require.Equal(t, math.NewInt(20), result[1].Amount.Amount)

	// newVal2 receives 20 from oldVal2 and 30 from oldVal3.
	require.Equal(t, newVal2.String(), result[2].ValidatorDstAddress)
	require.Equal(t, oldVal2.String(), result[2].ValidatorSrcAddress)
	require.Equal(t, newVal2.String(), result[2].ValidatorDstAddress)
	require.Equal(t, math.NewInt(20), result[2].Amount.Amount)
	require.Equal(t, oldVal3.String(), result[3].ValidatorSrcAddress)
	require.Equal(t, newVal2.String(), result[3].ValidatorDstAddress)
	require.Equal(t, math.NewInt(30), result[3].Amount.Amount)
}

// Total shares do not divide evenly: the last new validator absorbs the remainder.
func TestCalcRedelegations_UnevenTotal_LastValidatorAbsorbsRemainder(t *testing.T) {
	oldVal := makeValAddr(1)
	newVals := []sdk.ValAddress{makeValAddr(10), makeValAddr(11), makeValAddr(12)}
	// 100 / 3 = 33.333...
	expectedAmountPer := math.NewInt(100).Quo(math.NewInt(3))

	delegations := []stakingtypes.DelegationResponse{
		makeDelegation(oldVal, "100"),
	}
	result := calcRedelegationsHelper(delegations, newVals)

	require.Len(t, result, 3)

	require.Equal(t, expectedAmountPer, result[0].Amount.Amount)
	require.Equal(t, expectedAmountPer, result[1].Amount.Amount)

	// Last validator gets whatever remains after the first two took their exact share.
	expectedLast := math.NewInt(100).Sub(expectedAmountPer).Sub(expectedAmountPer)
	require.Equal(t, expectedLast, result[2].Amount.Amount)

	// Conservation: 99 tokens must be accounted for.
	require.Equal(t, math.NewInt(100), totalRedelegated(result))
}

// Source-validator amounts are fully consumed across a larger mixed scenario.
func TestCalcRedelegations_SourceAmountsAreFullyConsumed(t *testing.T) {
	oldVal1 := makeValAddr(1) // 70
	oldVal2 := makeValAddr(2) // 80
	oldVal3 := makeValAddr(3) // 50
	// total = 200, 4 new validators → 50 each (exact)
	newVals := []sdk.ValAddress{makeValAddr(10), makeValAddr(11), makeValAddr(12), makeValAddr(13)}
	expectedAmountPer := math.NewInt(50)

	delegations := []stakingtypes.DelegationResponse{
		makeDelegation(oldVal1, "70"),
		makeDelegation(oldVal2, "80"),
		makeDelegation(oldVal3, "50"),
	}
	result := calcRedelegationsHelper(delegations, newVals)

	for i := range newVals {
		require.Equal(t, expectedAmountPer, dstTokensFor(result, newVals[i].String()), "validator index %d", i)
	}

	require.Equal(t, math.NewInt(70), srcTokensFor(result, oldVal1.String()))
	require.Equal(t, math.NewInt(80), srcTokensFor(result, oldVal2.String()))
	require.Equal(t, math.NewInt(50), srcTokensFor(result, oldVal3.String()))
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

	delegations := make([]stakingtypes.DelegationResponse, len(specs))
	for i, s := range specs {
		delegations[i] = makeDelegation(s.addr, s.shares)
	}

	result := calcRedelegationsHelper(delegations, newVals)

	for i := range newVals {
		require.Equal(t, expectedAmountPer.TruncateInt(), dstTokensFor(result, newVals[i].String()), "new validator index %d should receive exactly %s shares", i, expectedAmountPer)
	}

	require.Equal(t, math.NewInt(600), totalRedelegated(result))

	// Every old validator's shares must be fully redistributed.
	for _, s := range specs {
		expected, _ := math.NewIntFromString(s.shares)
		require.Equal(t, expected, srcTokensFor(result, s.addr.String()), "old validator %s shares not fully consumed", s.addr)
	}
}

// Every message must carry the correct delegator address and denom.
func TestCalcRedelegations_MsgFieldsAreFullyPopulated(t *testing.T) {
	oldVal := makeValAddr(1)
	newVals := []sdk.ValAddress{makeValAddr(10), makeValAddr(11)}

	delegations := []stakingtypes.DelegationResponse{
		makeDelegation(oldVal, "100"),
	}
	result := calcRedelegationsHelper(delegations, newVals)

	for _, r := range result {
		require.Equal(t, testDelegator, r.DelegatorAddress)
		require.Equal(t, appparams.DefaultDenom, r.Amount.Denom)
		require.False(t, r.Amount.Amount.IsZero())
	}
}

// Destination in each MsgBeginRedelegate must match the enclosing Redelegation.ValidatorAddress.
func TestCalcRedelegations_DstAddressMatchesRedelegationValidator(t *testing.T) {
	oldVals := []sdk.ValAddress{makeValAddr(1), makeValAddr(2)}
	newVals := []sdk.ValAddress{makeValAddr(10), makeValAddr(11)}

	delegations := []stakingtypes.DelegationResponse{
		makeDelegation(oldVals[0], "50"),
		makeDelegation(oldVals[1], "50"),
	}
	result := calcRedelegationsHelper(delegations, newVals)

	for _, r := range result {
		require.Equal(t, r.ValidatorDstAddress, r.ValidatorDstAddress,
			"ValidatorDstAddress must equal the enclosing Redelegation.ValidatorAddress")
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

	delegations := []stakingtypes.DelegationResponse{
		makeDelegation(oldVals[0], "50"),
		makeDelegation(oldVals[1], "50"),
	}
	result := calcRedelegationsHelper(delegations, newVals)

	for _, r := range result {
		_, isOld := oldValSet[r.ValidatorSrcAddress]
		require.True(t, isOld, "ValidatorSrcAddress %s is not an old validator", r.ValidatorSrcAddress)
	}
}

func TestCalcRedelegations_OldNewMixed(t *testing.T) {
	oldVals := []sdk.ValAddress{makeValAddr(1), makeValAddr(2)}
	newVals := []sdk.ValAddress{makeValAddr(1), makeValAddr(11), makeValAddr(12), makeValAddr(13)}

	delegations := []stakingtypes.DelegationResponse{
		makeDelegation(oldVals[0], "50"),
		makeDelegation(oldVals[1], "50"),
	}
	result := calcRedelegationsHelper(delegations, newVals)

	totalMsgs := 0
	for _, r := range result {
		totalMsgs += 1
		isFirst := r.ValidatorDstAddress == makeValAddr(1).String()
		require.False(t, isFirst, "ValidatorDstAddress %s should not be redelegation message receiver", r.ValidatorDstAddress)
	}
	require.Equal(t, 3, totalMsgs)
}

func TestCalcRedelegations_OldNewTheSame(t *testing.T) {
	oldVals := []sdk.ValAddress{makeValAddr(1), makeValAddr(2)}
	newVals := []sdk.ValAddress{makeValAddr(1), makeValAddr(2)}

	delegations := []stakingtypes.DelegationResponse{
		makeDelegation(oldVals[0], "50"),
		makeDelegation(oldVals[1], "50"),
	}
	result := calcRedelegationsHelper(delegations, newVals)

	totalMsgs := len(result)
	// do nothing
	require.Equal(t, 0, totalMsgs)

	delegations = []stakingtypes.DelegationResponse{
		makeDelegation(oldVals[0], "30"),
		makeDelegation(oldVals[1], "70"),
	}
	result = calcRedelegationsHelper(delegations, newVals)

	totalMsgs = len(result)

	// single message to redistribute 20shares from oldVals[1] to oldVals[0]
	msg := result[0]
	require.Equal(t, oldVals[1].String(), msg.ValidatorSrcAddress)
	require.Equal(t, oldVals[0].String(), msg.ValidatorDstAddress)
	require.Equal(t, math.NewInt(20), msg.Amount.Amount)
	require.Equal(t, 1, totalMsgs)
}

func TestCalcRedelegations_NewIsSubSetOfOld(t *testing.T) {
	oldVals := []sdk.ValAddress{makeValAddr(1), makeValAddr(2), makeValAddr(3), makeValAddr(4)}
	newVals := []sdk.ValAddress{makeValAddr(1), makeValAddr(4)}

	delegations := []stakingtypes.DelegationResponse{
		makeDelegation(oldVals[0], "50"),
		makeDelegation(oldVals[1], "50"),
		makeDelegation(oldVals[2], "50"),
		makeDelegation(oldVals[3], "50"),
	}
	result := calcRedelegationsHelper(delegations, newVals)

	totalMsgs := len(result)
	msg1 := result[0]
	require.Equal(t, oldVals[1].String(), msg1.ValidatorSrcAddress)
	require.Equal(t, oldVals[0].String(), msg1.ValidatorDstAddress)
	require.Equal(t, math.NewInt(50), msg1.Amount.Amount)
	msg2 := result[1]
	require.Equal(t, oldVals[2].String(), msg2.ValidatorSrcAddress)
	require.Equal(t, oldVals[3].String(), msg2.ValidatorDstAddress)
	require.Equal(t, math.NewInt(50), msg2.Amount.Amount)
	require.Equal(t, 2, totalMsgs)
}

func TestCalcRedelegations_NewIsSource(t *testing.T) {
	oldVals := []sdk.ValAddress{makeValAddr(1), makeValAddr(2), makeValAddr(3), makeValAddr(4)}
	newVals := []sdk.ValAddress{makeValAddr(1), makeValAddr(4)}

	delegations := []stakingtypes.DelegationResponse{
		makeDelegation(oldVals[0], "150"),
		makeDelegation(oldVals[1], "10"),
		makeDelegation(oldVals[2], "10"),
		makeDelegation(oldVals[3], "10"),
	}
	result := calcRedelegationsHelper(delegations, newVals)

	totalMsgs := len(result)
	msg1 := result[0]
	require.Equal(t, oldVals[0].String(), msg1.ValidatorSrcAddress)
	require.Equal(t, oldVals[3].String(), msg1.ValidatorDstAddress)
	require.Equal(t, math.NewInt(60), msg1.Amount.Amount)
	msg2 := result[1]
	require.Equal(t, oldVals[1].String(), msg2.ValidatorSrcAddress)
	require.Equal(t, oldVals[3].String(), msg2.ValidatorDstAddress)
	require.Equal(t, math.NewInt(10), msg2.Amount.Amount)
	msg3 := result[2]
	require.Equal(t, oldVals[2].String(), msg3.ValidatorSrcAddress)
	require.Equal(t, oldVals[3].String(), msg3.ValidatorDstAddress)
	require.Equal(t, math.NewInt(10), msg3.Amount.Amount)
	require.Equal(t, 3, totalMsgs)
}
