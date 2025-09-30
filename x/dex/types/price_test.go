package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	math_utils "github.com/neutron-org/neutron/v8/utils/math"
	"github.com/neutron-org/neutron/v8/x/dex/types"
)

func TestCalcTickIndexFromPrice(t *testing.T) {
	for _, tc := range []struct {
		desc string
		tick int64
		err  error
	}{
		{
			desc: "0",
			tick: 0,
		},
		{
			desc: "10",
			tick: 10,
		},
		{
			desc: "-10",
			tick: -10,
		},
		{
			desc: "100000",
			tick: 100000,
		},
		{
			desc: "-100000",
			tick: -100000,
		},
		{
			desc: "-200000",
			tick: -200000,
		},
		{
			desc: "-400000",
			tick: -400000,
		},
		{
			desc: "MaxTickExp",
			tick: int64(types.MaxTickExp),
		},
		{
			desc: "MinTickExp",
			tick: int64(types.MaxTickExp) * -1,
		},
		{
			desc: "GT MaxTickExp",
			tick: int64(types.MaxTickExp) + 1,
			err:  types.ErrTickOutsideRange,
		},
		{
			desc: "LT MinTickExp",
			tick: -1*int64(types.MaxTickExp) - 1,
			err:  types.ErrTickOutsideRange,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			price, err1 := types.CalcPrice(tc.tick)
			val, err2 := types.CalcTickIndexFromPrice(price)
			if err1 != nil {
				require.ErrorIs(t, err1, tc.err)
				require.ErrorIs(t, err2, tc.err)
			} else {
				// If we are not outside the tick range we should TestCalcTickIndexFromPrice to never throw
				require.NoError(t, err2)
				require.Equal(t, tc.tick, val)
			}
		})
	}
}
func TestPriceDups(t *testing.T) {
	prevPrice := math_utils.ZeroPrecDec()
	for i := 0; i >= int(types.MaxTickExp)*-1; i-- {
		price, err := types.CalcPrice(int64(i))
		require.NoError(t, err)
		if price.Equal(prevPrice) {
			t.Fatalf("Price (%v) %s is equal to previous price %s", i, price, prevPrice)
		}
		prevPrice = price
	}
}
