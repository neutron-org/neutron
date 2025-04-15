package utils_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
	"github.com/neutron-org/neutron/v6/x/dex/utils"
)

func TestParsePrecDecScientificNotation(t *testing.T) {
	for _, tc := range []struct {
		desc string
		in   string
		out  math_utils.PrecDec
		err  bool
	}{
		{
			desc: "invalid garbage",
			in:   "sdfdsf",
			err:  true,
		},
		{
			desc: "invalid exp",
			in:   "10Z-5",
			err:  true,
		},
		{
			desc: "invalid exp 2",
			in:   "10E-",
			err:  true,
		},
		{
			desc: "invalid exp 3",
			in:   "10E-+1",
			err:  true,
		},
		{
			desc: "invalid exp 4",
			in:   "10E+1.0",
			err:  true,
		},
		{
			desc: "valid integer",
			in:   "10",
			out:  math_utils.NewPrecDec(10),
		},
		{
			desc: "valid decimal",
			in:   "10.1",
			out:  math_utils.MustNewPrecDecFromStr("10.1"),
		},
		{
			desc: "valid integer with E-6",
			in:   "6E-6",
			out:  math_utils.MustNewPrecDecFromStr("0.000006"),
		},
		{
			desc: "valid decimal with E+0",
			in:   "10.1E+0",
			out:  math_utils.MustNewPrecDecFromStr("10.1"),
		},
		{
			desc: "valid decimal with E+1",
			in:   "10.1E+1",
			out:  math_utils.MustNewPrecDecFromStr("101"),
		},
		{
			desc: "valid decimal with E+10",
			in:   "10.12345678901E+10",
			out:  math_utils.MustNewPrecDecFromStr("101234567890.1"),
		},
		{
			desc: "valid decimal with E-0",
			in:   "0.1E-0",
			out:  math_utils.MustNewPrecDecFromStr("0.1"),
		},
		{
			desc: "valid decimal with E-1",
			in:   "189.3E-1",
			out:  math_utils.MustNewPrecDecFromStr("18.93"),
		},
		{
			desc: "valid decimal with E-10",
			in:   "101234567890.1E-10",
			out:  math_utils.MustNewPrecDecFromStr("10.12345678901"),
		},
		{
			desc: "valid with lowercase 'e'",
			in:   "10.1e-2",
			out:  math_utils.MustNewPrecDecFromStr("0.101"),
		},
		{
			desc: "valid with no '+'",
			in:   "10.1E3",
			out:  math_utils.MustNewPrecDecFromStr("10100"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			val, err := utils.ParsePrecDecScientificNotation(tc.in)
			if tc.err {
				require.Error(t, err)
			} else {
				require.True(t, val.Equal(tc.out), "Got: %v; Expected: %v", val, tc.out)
			}
		})
	}
}
