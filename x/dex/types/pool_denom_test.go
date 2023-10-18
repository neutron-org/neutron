package types_test

import (
	"testing"

	. "github.com/neutron-org/neutron/x/dex/types"
	"github.com/stretchr/testify/require"
)

func TestParsePoolIDFromDenom(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		denom    string
		expected uint64
		valid    bool
	}{
		{
			desc:     "valid denom",
			denom:    "neutron/pool/0",
			expected: 0,
			valid:    true,
		},
		{
			desc:     "valid denom long",
			denom:    "neutron/pool/999999999",
			expected: 999999999,
			valid:    true,
		},
		{
			desc:  "invalid format 1",
			denom: "/neutron/pool/0",
			valid: false,
		},
		{
			desc:  "invalid format 2",
			denom: "neutron/pool/0/",
			valid: false,
		},
		{
			desc:  "invalid denom prefix",
			denom: "INVALID/pool/0",
			valid: false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			id, err := ParsePoolIDFromDenom(tc.denom)
			if tc.valid {
				require.NoError(t, err)
				require.Equal(t, tc.expected, id)
			} else {
				require.Error(t, err)
			}
		})
	}
}
