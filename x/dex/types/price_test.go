package types_test

// This will continue to fail until we upgrade away from sdk.Dec
// func TestPriceMath(t *testing.T) {
// 	tick := 352437
// 	amount := sdk.MustNewDecFromStr("1000000000000000000000")
// 	basePrice := utils.BasePrice()
// 	expected := amount.Quo(basePrice.Power(uint64(tick))).TruncateInt()
// 	result := types.MustCalcPrice(int64(tick)).Mul(amount).TruncateInt()
// 	assert.Equal(t, expected.Int64(), result.Int64())
// }
