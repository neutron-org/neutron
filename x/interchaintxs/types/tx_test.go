package types_test

import (
	"testing"

	cosmosTypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/x/interchaintxs/types"
)

const TestAddress = "neutron1m9l358xunhhwds0568za49mzhvuxx9ux8xafx2"

func TestMsgRegisterInterchainAccountGetSigners(t *testing.T) {
	tests := []struct {
		name     string
		malleate func() sdktypes.LegacyMsg
	}{
		{
			"valid_signer",
			func() sdktypes.LegacyMsg {
				return &types.MsgRegisterInterchainAccount{
					FromAddress:         TestAddress,
					ConnectionId:        "connection-id",
					InterchainAccountId: "1",
				}
			},
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()
		addr, _ := sdktypes.AccAddressFromBech32(TestAddress)
		require.Equal(t, msg.GetSigners(), []sdktypes.AccAddress{addr})
	}
}

func TestMsgSubmitTXGetSigners(t *testing.T) {
	tests := []struct {
		name     string
		malleate func() sdktypes.LegacyMsg
	}{
		{
			"valid_signer",
			func() sdktypes.LegacyMsg {
				return &types.MsgSubmitTx{
					FromAddress:         TestAddress,
					ConnectionId:        "connection-id",
					InterchainAccountId: "1",
					Msgs: []*cosmosTypes.Any{{
						TypeUrl: "msg",
						Value:   []byte{100}, // just check that values are not nil
					}},
				}
			},
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()
		addr, _ := sdktypes.AccAddressFromBech32(TestAddress)
		require.Equal(t, msg.GetSigners(), []sdktypes.AccAddress{addr})
	}
}
