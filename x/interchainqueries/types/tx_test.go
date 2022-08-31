package types_test

import (
	"testing"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/proto/tendermint/crypto"

	iqtypes "github.com/neutron-org/neutron/x/interchainqueries/types"
)

const TestAddress = "cosmos10h9stc5v6ntgeygf5xf945njqq5h32r53uquvw"

func TestMsgRegisterInterchainQueryValidate(t *testing.T) {
	tests := []struct {
		name        string
		malleate    func() sdktypes.Msg
		expectedErr error
	}{
		{
			"invalid query type",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          "invalid_type",
					ZoneId:             "id",
					UpdatePeriod:       1,
					Sender:             TestAddress,
				}
			},
			iqtypes.ErrInvalidQueryType,
		},
		{
			"invalid transactions filter format",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "&)(^Y(*&(*&(&(*",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					ZoneId:             "id",
					UpdatePeriod:       1,
					Sender:             TestAddress,
				}
			},
			iqtypes.ErrInvalidTransactionsFilter,
		},
		{
			"invalid update period",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					ZoneId:             "osmosis",
					UpdatePeriod:       0,
					Sender:             TestAddress,
				}
			},
			iqtypes.ErrInvalidUpdatePeriod,
		},
		{
			"empty sender",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					ZoneId:             "osmosis",
					UpdatePeriod:       1,
					Sender:             "",
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid sender",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					ZoneId:             "osmosis",
					UpdatePeriod:       1,
					Sender:             "cosmos14234_invalid_address",
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"empty connection id",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "",
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					ZoneId:             "osmosis",
					UpdatePeriod:       1,
					Sender:             TestAddress,
				}
			},
			iqtypes.ErrInvalidConnectionID,
		},
		{
			"invalid zone id",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					ZoneId:             "",
					UpdatePeriod:       1,
					Sender:             TestAddress,
				}
			},
			iqtypes.ErrInvalidZoneID,
		},
		{
			"valid",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeKV),
					ZoneId:             "osmosis",
					UpdatePeriod:       1,
					Sender:             TestAddress,
				}
			},
			nil,
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()

		if tt.expectedErr != nil {
			require.ErrorIs(t, msg.ValidateBasic(), tt.expectedErr)
		} else {
			require.NoError(t, msg.ValidateBasic())
		}
	}
}

func TestMsgSubmitQueryResultValidate(t *testing.T) {
	tests := []struct {
		name        string
		malleate    func() sdktypes.Msg
		expectedErr error
	}{
		{
			"valid",
			func() sdktypes.Msg {
				return &iqtypes.MsgSubmitQueryResult{
					QueryId:  1,
					Sender:   TestAddress,
					ClientId: "client-id",
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
							StoragePrefix: host.StoreKey,
						}},
						Block:    nil,
						Height:   100,
						Revision: 1,
					},
				}
			},
			nil,
		},
		{
			"empty result",
			func() sdktypes.Msg {
				return &iqtypes.MsgSubmitQueryResult{
					QueryId:  1,
					Sender:   TestAddress,
					ClientId: "client-id",
					Result:   nil,
				}
			},
			iqtypes.ErrEmptyResult,
		},
		{
			"empty kv results and block result",
			func() sdktypes.Msg {
				return &iqtypes.MsgSubmitQueryResult{
					QueryId:  1,
					Sender:   TestAddress,
					ClientId: "client-id",
					Result: &iqtypes.QueryResult{
						KvResults: nil,
						Block:     nil,
						Height:    100,
						Revision:  1,
					},
				}
			},
			iqtypes.ErrEmptyResult,
		},
		{
			"zero query id",
			func() sdktypes.Msg {
				return &iqtypes.MsgSubmitQueryResult{
					QueryId:  0,
					Sender:   TestAddress,
					ClientId: "client-id",
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
							StoragePrefix: host.StoreKey,
						}},
						Block:    nil,
						Height:   100,
						Revision: 1,
					},
				}
			},
			iqtypes.ErrInvalidQueryID,
		},
		{
			"empty sender",
			func() sdktypes.Msg {
				return &iqtypes.MsgSubmitQueryResult{
					QueryId:  1,
					Sender:   "",
					ClientId: "client-id",
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
							StoragePrefix: host.StoreKey,
						}},
						Block:    nil,
						Height:   100,
						Revision: 1,
					},
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid sender",
			func() sdktypes.Msg {
				return &iqtypes.MsgSubmitQueryResult{
					QueryId:  1,
					Sender:   "invalid_sender",
					ClientId: "client-id",
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
							StoragePrefix: host.StoreKey,
						}},
						Block:    nil,
						Height:   100,
						Revision: 1,
					},
				}
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"empty client id",
			func() sdktypes.Msg {
				return &iqtypes.MsgSubmitQueryResult{
					QueryId:  1,
					Sender:   TestAddress,
					ClientId: "",
					Result: &iqtypes.QueryResult{
						KvResults: nil,
						Block: &iqtypes.Block{
							NextBlockHeader: nil,
							Header:          nil,
							Tx:              nil,
						},
						Height:   100,
						Revision: 1,
					},
				}
			},
			iqtypes.ErrInvalidClientID,
		},
	}

	for _, tt := range tests {
		msg := tt.malleate()

		if tt.expectedErr != nil {
			require.ErrorIs(t, msg.ValidateBasic(), tt.expectedErr)
		} else {
			require.NoError(t, msg.ValidateBasic())
		}
	}
}

func TestMsgRegisterInterchainQueryGetSigners(t *testing.T) {
	tests := []struct {
		name     string
		malleate func() sdktypes.Msg
	}{
		{
			"valid_signer",
			func() sdktypes.Msg {
				return &iqtypes.MsgRegisterInterchainQuery{
					ConnectionId:       "connection-0",
					TransactionsFilter: "{}",
					Keys:               nil,
					QueryType:          string(iqtypes.InterchainQueryTypeTX),
					ZoneId:             "osmosis",
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
		malleate func() sdktypes.Msg
	}{
		{
			"valid_signer",
			func() sdktypes.Msg {
				return &iqtypes.MsgSubmitQueryResult{
					QueryId:  1,
					Sender:   TestAddress,
					ClientId: "client-id",
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
							StoragePrefix: host.StoreKey,
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
