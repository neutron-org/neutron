package types_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func TestCalcTickIndexFromPrice(t *testing.T) {
	for _, tc := range []struct {
		desc string
		tick int64
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
			desc: "-100000",
			tick: -100000,
		},
		{
			desc: "-100000",
			tick: -100000,
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
		},
		{
			desc: "LT MinTickExp",
			tick: -1*int64(types.MaxTickExp) - 1,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			price, err1 := types.CalcPrice(tc.tick)
			val, err2 := types.CalcTickIndexFromPrice(price)
			if errors.Is(err1, types.ErrTickOutsideRange) {
				require.ErrorIs(t, err2, types.ErrPriceOutsideRange)
			} else {
				// Only expected error is ErrTickOutsideRange.
				// If we are not outside the tick range we should TestCalcTickIndexFromPrice to never throw
				require.NoError(t, err2)
				require.Equal(t, tc.tick, val)
			}
		})
	}
}
