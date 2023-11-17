package types_test

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/math"
	forwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	"github.com/iancoleman/orderedmap"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/app"
	"github.com/neutron-org/neutron/testutil/common/sample"
	"github.com/neutron-org/neutron/x/dex/types"
	. "github.com/neutron-org/neutron/x/ibcswap/types"
)

func init() {
	_ = app.GetDefaultConfig()
}

// TestPacketMetadata_Marshal asserts that the marshaling of the swap metadata works as intended.
func TestPacketMetadata_Marshal(t *testing.T) {
	pm := PacketMetadata{
		&SwapMetadata{
			MsgPlaceLimitOrder: &types.MsgPlaceLimitOrder{
				Creator:          "test-1",
				Receiver:         "test-1",
				TokenIn:          "token-a",
				TokenOut:         "token-b",
				AmountIn:         math.NewInt(123),
				TickIndexInToOut: 0,
				OrderType:        types.LimitOrderType_FILL_OR_KILL,
			},
			Next: nil,
		},
	}
	_, err := json.Marshal(pm)
	require.NoError(t, err)
}

// TestPacketMetadata_MarshalWithNext asserts that the marshaling of the swap metadata works as intended with next field initialized.
func TestPacketMetadata_MarshalWithNext(t *testing.T) {
	forwardMedata := &forwardtypes.PacketMetadata{
		Forward: &forwardtypes.ForwardMetadata{
			Receiver: "cosmos14zde8usc4ur04y3aqnufzzmv2uqdpwwttr5uwv",
			Port:     "transfer",
			Channel:  "channel-0",
			Timeout:  0,
			Retries:  nil,
			Next:     nil,
		},
	}
	nextBz, err := json.Marshal(forwardMedata)
	require.NoError(t, err)

	pm := PacketMetadata{
		&SwapMetadata{
			MsgPlaceLimitOrder: &types.MsgPlaceLimitOrder{
				Creator:          "test-1",
				Receiver:         "test-1",
				TokenIn:          "token-a",
				TokenOut:         "token-b",
				TickIndexInToOut: 0,
				AmountIn:         math.NewInt(123),
				OrderType:        types.LimitOrderType_FILL_OR_KILL,
				// MaxAmountOut: math.NewInt(456),
			},
			Next: NewJSONObject(false, nextBz, orderedmap.OrderedMap{}),
		},
	}
	_, err = json.Marshal(pm)
	require.NoError(t, err)
}

// TestPacketMetadata_Unmarshal asserts that unmarshaling works as intended.
func TestPacketMetadata_Unmarshal(t *testing.T) {
	metadata := "{\n  \"swap\": {\n    \"creator\": \"test-1\",\n \"TickIndexInToOut\": 0,\n \"orderType\": 1,\n   \"receiver\": \"test-1\",\n    \"tokenIn\": \"token-a\",\n    \"tokenOut\": \"token-b\",\n    \"AmountIn\": \"123\",\n    \"next\": \"\"\n  }\n}"
	pm := &PacketMetadata{}
	err := json.Unmarshal([]byte(metadata), pm)
	require.NoError(t, err)
}

// TestPacketMetadata_UnmarshalStringNext asserts that unmarshaling works as intended when next is escaped json string.
func TestPacketMetadata_UnmarshalStringNext(t *testing.T) {
	metadata := "{\n  \"swap\": {\n    \"creator\": \"test-1\",\n    \"receiver\": \"test-1\",\n    \"tokenIn\": \"token-a\",\n    \"tokenOut\": \"token-b\",\n    \"AmountIn\": \"123\",\n  \"TickIndexInToOut\": 0,\n \"orderType\": 1,\n  \"next\": \" {\\\"forward\\\":{\\\"receiver\\\":\\\"cosmos1f4cur2krsua2th9kkp7n0zje4stea4p9tu70u8\\\",\\\"port\\\":\\\"transfer\\\",\\\"channel\\\":\\\"channel-0\\\",\\\"timeout\\\":0,\\\"next\\\":{\\\"forward\\\":{\\\"receiver\\\":\\\"cosmos1l505zhahp24v5jsmps9vs5asah759fdce06sfp\\\",\\\"port\\\":\\\"transfer\\\",\\\"channel\\\":\\\"channel-0\\\",\\\"timeout\\\":0}}}}\"\n  }\n}"
	pm := &PacketMetadata{}
	err := json.Unmarshal([]byte(metadata), pm)
	require.NoError(t, err)
}

// TestPacketMetadata_UnmarshalJSONNext asserts that unmarshaling works as intended when next is a raw json object.
func TestPacketMetadata_UnmarshalJSONNext(t *testing.T) {
	metadata := "{\"swap\":{\"creator\":\"test-1\",\"receiver\":\"test-1\",\"tokenIn\":\"token-a\",\"tokenOut\":\"token-b\",\"AmountIn\":\"123\",\"TickIndexInToOut\":0, \"orderType\": 1, \"tokenIn\":\"token-in\",\"next\":{\"forward\":{\"receiver\":\"cosmos14zde8usc4ur04y3aqnufzzmv2uqdpwwttr5uwv\",\"port\":\"transfer\",\"channel\":\"channel-0\"}}}}"
	pm := &PacketMetadata{}
	err := json.Unmarshal([]byte(metadata), pm)
	require.NoError(t, err)
}

func TestSwapMetadata_ValidatePass(t *testing.T) {
	pm := PacketMetadata{
		&SwapMetadata{
			MsgPlaceLimitOrder: &types.MsgPlaceLimitOrder{
				Creator:          sample.AccAddress(),
				Receiver:         sample.AccAddress(),
				TokenIn:          "token-a",
				TokenOut:         "token-b",
				AmountIn:         math.NewInt(123),
				TickIndexInToOut: 0,
				OrderType:        types.LimitOrderType_FILL_OR_KILL,
			},
			Next: nil,
		},
	}
	_, err := json.Marshal(pm)
	require.NoError(t, err)

	require.NoError(t, pm.Swap.Validate())
}

func TestSwapMetadata_ValidateFail(t *testing.T) {
	pm := PacketMetadata{
		&SwapMetadata{
			MsgPlaceLimitOrder: &types.MsgPlaceLimitOrder{
				Creator:          "",
				Receiver:         "test-1",
				TokenIn:          "token-a",
				TokenOut:         "token-b",
				AmountIn:         math.NewInt(123),
				TickIndexInToOut: 0,
				OrderType:        types.LimitOrderType_FILL_OR_KILL,
			},
			Next: nil,
		},
	}
	_, err := json.Marshal(pm)
	require.NoError(t, err)
	require.Error(t, pm.Swap.Validate())

	pm = PacketMetadata{
		&SwapMetadata{
			MsgPlaceLimitOrder: &types.MsgPlaceLimitOrder{
				Creator:          "creator",
				Receiver:         "",
				TokenIn:          "token-a",
				TokenOut:         "token-b",
				AmountIn:         math.NewInt(123),
				TickIndexInToOut: 0,
				OrderType:        types.LimitOrderType_FILL_OR_KILL,
			},
			Next: nil,
		},
	}
	_, err = json.Marshal(pm)
	require.NoError(t, err)
	require.Error(t, pm.Swap.Validate())

	pm = PacketMetadata{
		&SwapMetadata{
			MsgPlaceLimitOrder: &types.MsgPlaceLimitOrder{
				Creator:          "creator",
				Receiver:         "test-1",
				TokenIn:          "",
				TokenOut:         "token-b",
				AmountIn:         math.NewInt(123),
				TickIndexInToOut: 0,
				OrderType:        types.LimitOrderType_FILL_OR_KILL,
			},
			Next: nil,
		},
	}
	_, err = json.Marshal(pm)
	require.NoError(t, err)
	require.Error(t, pm.Swap.Validate())

	pm = PacketMetadata{
		&SwapMetadata{
			MsgPlaceLimitOrder: &types.MsgPlaceLimitOrder{
				Creator:          "creator",
				Receiver:         "receiver",
				TokenIn:          "token-a",
				TokenOut:         "",
				AmountIn:         math.NewInt(123),
				TickIndexInToOut: 0,
				OrderType:        types.LimitOrderType_FILL_OR_KILL,
			},
			Next: nil,
		},
	}
	_, err = json.Marshal(pm)
	require.NoError(t, err)
	require.Error(t, pm.Swap.Validate())

	pm = PacketMetadata{
		&SwapMetadata{
			MsgPlaceLimitOrder: &types.MsgPlaceLimitOrder{
				Creator:          "creator",
				Receiver:         "receiver",
				TokenIn:          "token-a",
				TokenOut:         "token-b",
				AmountIn:         math.NewInt(0),
				TickIndexInToOut: 0,
				OrderType:        types.LimitOrderType_FILL_OR_KILL,
			},
			Next: nil,
		},
	}
	_, err = json.Marshal(pm)
	require.NoError(t, err)
	require.Error(t, pm.Swap.Validate())

	pm = PacketMetadata{
		&SwapMetadata{
			MsgPlaceLimitOrder: &types.MsgPlaceLimitOrder{
				Creator:          "creator",
				Receiver:         "receiver",
				TokenIn:          "token-a",
				TokenOut:         "token-b",
				AmountIn:         math.NewInt(-1),
				TickIndexInToOut: 0,
				OrderType:        types.LimitOrderType_FILL_OR_KILL,
			},
			Next: nil,
		},
	}
	_, err = json.Marshal(pm)
	require.NoError(t, err)
	require.Error(t, pm.Swap.Validate())

	pm = PacketMetadata{
		&SwapMetadata{
			MsgPlaceLimitOrder: &types.MsgPlaceLimitOrder{
				Creator:          "creator",
				Receiver:         "receiver",
				TokenIn:          "token-a",
				TokenOut:         "token-b",
				AmountIn:         math.NewInt(123),
				TickIndexInToOut: 0,
				OrderType:        types.LimitOrderType_GOOD_TIL_CANCELLED,
			},
			Next: nil,
		},
	}
	_, err = json.Marshal(pm)
	require.NoError(t, err)
	require.Error(t, pm.Swap.Validate())
}
