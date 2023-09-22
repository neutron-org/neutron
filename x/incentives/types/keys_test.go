package types_test

import (
	"bytes"
	"testing"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/neutron-org/neutron/x/incentives/types"
	"github.com/stretchr/testify/require"
)

func TestGetTimeKey(t *testing.T) {
	now := time.Now()
	timeKey := GetTimeKey(now)
	require.True(t, bytes.HasPrefix(timeKey, KeyPrefixTimestamp))
	require.True(t, bytes.HasSuffix(timeKey, sdk.FormatTimeBytes(now)))
}
