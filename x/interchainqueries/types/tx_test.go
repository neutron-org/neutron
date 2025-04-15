package types_test

import (
	"testing"

	"github.com/cometbft/cometbft/proto/tendermint/crypto"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	ibchost "github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/stretchr/testify/require"

	iqtypes "github.com/neutron-org/neutron/v6/x/interchainqueries/types"
)

const TestAddress = "cosmos10h9stc5v6ntgeygf5xf945njqq5h32r53uquvw"

func TestMsgRegisterInterchainQueryGetSigners(t *testing.T) {
	tests := []struct {
		name     string
		malleate func() sdktypes.LegacyMsg
	}{
		{
			"valid_signer",
			func() sdktypes.LegacyMsg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					UpdatePeriod:       1,
					Sender:             TestAddress,
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

func TestMsgSubmitQueryResultGetSigners(t *testing.T) {
	tests := []struct {
		name     string
		malleate func() sdktypes.LegacyMsg
	}{
		{
			"valid_signer",
			func() sdktypes.LegacyMsg {
				return &iqtypes.MsgSubmitQueryResult{
					QueryId: 1,
					Sender:  TestAddress,
					Result: &iqtypes.QueryResult{
						KvResults: []*iqtypes.StorageValue{{
							Key: []byte{10},
							Proof: &crypto.ProofOps{Ops: []crypto.ProofOp{
								{
									Type: "type",
									Key:  []byte{10},
									Data: []byte{10},
								},
							}},
							Value:         []byte{10},
							StoragePrefix: ibchost.StoreKey,
						}},
						Block:    nil,
						Height:   100,
						Revision: 1,
					},
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

func TestMsgUpdateQueryGetSigners(t *testing.T) {
	tests := []struct {
		name     string
		malleate func() sdktypes.LegacyMsg
	}{
		{
			"valid_signer",
			func() sdktypes.LegacyMsg {
				return &iqtypes.MsgUpdateInterchainQueryRequest{
					Sender: TestAddress,
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

func TestMsgRemoveQueryGetSigners(t *testing.T) {
	tests := []struct {
		name     string
		malleate func() sdktypes.LegacyMsg
	}{
		{
			"valid_signer",
			func() sdktypes.LegacyMsg {
				return &iqtypes.MsgRemoveInterchainQueryRequest{
					Sender: TestAddress,
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
