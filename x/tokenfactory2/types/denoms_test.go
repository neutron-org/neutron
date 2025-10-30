package types_test

import (
	"strings"
	"testing"

	"github.com/neutron-org/neutron/v8/app/config"

	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v8/x/tokenfactory2/types"
)

func TestDecomposeDenoms(t *testing.T) {
	config.GetDefaultConfig()
	normalDenom, err := types.GetTokenDenom("neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2", "bitcoin")
	require.NoError(t, err)

	for _, tc := range []struct {
		desc  string
		denom string
		valid bool
	}{
		{
			desc:  "empty is invalid",
			denom: "",
			valid: false,
		},
		{
			desc:  "normal",
			denom: normalDenom,
			valid: true,
		},
		{
			desc:  "multiple separators in subdenom",
			denom: normalDenom + types.Separator + "1",
			valid: true,
		},
		{
			desc:  "no subdenom",
			denom: strings.SplitAfterN(normalDenom, types.Separator, 1)[0],
			valid: true,
		},
		{
			desc:  "incorrect prefix",
			denom: strings.ReplaceAll(normalDenom, types.ModuleDenomPrefix, "ibc"),
			valid: false,
		},
		{
			desc:  "subdenom of only separators",
			denom: normalDenom + types.Separator + types.Separator + types.Separator,
			valid: true,
		},
		{
			desc:  "too long name",
			denom: normalDenom + strings.Repeat("a", 500),
			valid: false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			_, _, err := types.DeconstructDenom(tc.denom)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
