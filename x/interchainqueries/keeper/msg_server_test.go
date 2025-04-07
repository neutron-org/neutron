package keeper_test

import (
	"testing"

	"github.com/cometbft/cometbft/proto/tendermint/crypto"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibchost "github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/testutil"
	testkeeper "github.com/neutron-org/neutron/v6/testutil/interchainqueries/keeper"
	"github.com/neutron-org/neutron/v6/x/interchainqueries/keeper"
	"github.com/neutron-org/neutron/v6/x/interchainqueries/types"
)

func TestMsgRegisterInterchainQueryValidate(t *testing.T) {
	k, ctx := testkeeper.InterchainQueriesKeeper(t, nil, nil, nil, nil)
	msgServer := keeper.NewMsgServerImpl(*k)

	tests := []struct {
		name        string
		msg         types.MsgRegisterInterchainQuery
		expectedErr error
	}{
		{
			"invalid update period",
			types.MsgRegisterInterchainQuery{
				QueryType:          string(types.InterchainQueryTypeTX),
				Keys:               nil,
				TransactionsFilter: "[]",
				ConnectionId:       "connection-0",
				UpdatePeriod:       0,
				Sender:             testutil.TestOwnerAddress,
			},
			types.ErrInvalidUpdatePeriod,
		},
		{
			"empty sender",
			types.MsgRegisterInterchainQuery{
				QueryType:          string(types.InterchainQueryTypeTX),
				Keys:               nil,
				TransactionsFilter: "[]",
				ConnectionId:       "connection-0",
				UpdatePeriod:       1,
				Sender:             "",
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid sender",
			types.MsgRegisterInterchainQuery{
				QueryType:          string(types.InterchainQueryTypeTX),
				Keys:               nil,
				TransactionsFilter: "[]",
				ConnectionId:       "connection-0",
				UpdatePeriod:       1,
				Sender:             "cosmos14234_invalid_address",
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"empty connection id",
			types.MsgRegisterInterchainQuery{
				QueryType:          string(types.InterchainQueryTypeTX),
				Keys:               nil,
				TransactionsFilter: "[]",
				ConnectionId:       "",
				UpdatePeriod:       1,
				Sender:             testutil.TestOwnerAddress,
			},
			types.ErrInvalidConnectionID,
		},
		{
			"invalid query type",
			types.MsgRegisterInterchainQuery{
				QueryType:          "invalid_type",
				Keys:               nil,
				TransactionsFilter: "[]",
				ConnectionId:       "connection-0",
				UpdatePeriod:       1,
				Sender:             testutil.TestOwnerAddress,
			},
			types.ErrInvalidQueryType,
		},
		{
			"empty keys",
			types.MsgRegisterInterchainQuery{
				QueryType:          string(types.InterchainQueryTypeKV),
				Keys:               nil,
				TransactionsFilter: "[]",
				ConnectionId:       "connection-0",
				UpdatePeriod:       1,
				Sender:             testutil.TestOwnerAddress,
			},
			types.ErrEmptyKeys,
		},
		{
			"too many keys",
			types.MsgRegisterInterchainQuery{
				QueryType:          string(types.InterchainQueryTypeKV),
				Keys:               make([]*types.KVKey, types.DefaultMaxKvQueryKeysCount+1),
				TransactionsFilter: "[]",
				ConnectionId:       "connection-0",
				UpdatePeriod:       1,
				Sender:             testutil.TestOwnerAddress,
			},
			types.ErrTooManyKVQueryKeys,
		},
		{
			"nil key",
			types.MsgRegisterInterchainQuery{
				QueryType:          string(types.InterchainQueryTypeKV),
				Keys:               []*types.KVKey{{Key: []byte("key1"), Path: "path1"}, nil},
				TransactionsFilter: "[]",
				ConnectionId:       "connection-0",
				UpdatePeriod:       1,
				Sender:             testutil.TestOwnerAddress,
			},
			sdkerrors.ErrInvalidType,
		},
		{
			"duplicated keys",
			types.MsgRegisterInterchainQuery{
				QueryType:          string(types.InterchainQueryTypeKV),
				Keys:               []*types.KVKey{{Key: []byte("key1"), Path: "path1"}, {Key: []byte("key1"), Path: "path1"}},
				TransactionsFilter: "[]",
				ConnectionId:       "connection-0",
				UpdatePeriod:       1,
				Sender:             testutil.TestOwnerAddress,
			},
			sdkerrors.ErrInvalidRequest,
		},
		{
			"empty key path",
			types.MsgRegisterInterchainQuery{
				QueryType:          string(types.InterchainQueryTypeKV),
				Keys:               []*types.KVKey{{Key: []byte("key1"), Path: ""}},
				TransactionsFilter: "[]",
				ConnectionId:       "connection-0",
				UpdatePeriod:       1,
				Sender:             testutil.TestOwnerAddress,
			},
			types.ErrEmptyKeyPath,
		},
		{
			"empty key id",
			types.MsgRegisterInterchainQuery{
				QueryType:          string(types.InterchainQueryTypeKV),
				Keys:               []*types.KVKey{{Key: []byte(""), Path: "path"}},
				TransactionsFilter: "[]",
				ConnectionId:       "connection-0",
				UpdatePeriod:       1,
				Sender:             testutil.TestOwnerAddress,
			},
			types.ErrEmptyKeyID,
		},
		{
			"invalid transactions filter format",
			types.MsgRegisterInterchainQuery{
				QueryType:          string(types.InterchainQueryTypeTX),
				Keys:               nil,
				TransactionsFilter: "&)(^Y(*&(*&(&(*",
				ConnectionId:       "connection-0",
				UpdatePeriod:       1,
				Sender:             testutil.TestOwnerAddress,
			},
			types.ErrInvalidTransactionsFilter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.RegisterInterchainQuery(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestMsgSubmitQueryResultValidate(t *testing.T) {
	k, ctx := testkeeper.InterchainQueriesKeeper(t, nil, nil, nil, nil)
	msgServer := keeper.NewMsgServerImpl(*k)

	tests := []struct {
		name        string
		msg         types.MsgSubmitQueryResult
		expectedErr error
	}{
		{
			"empty result",
			types.MsgSubmitQueryResult{
				QueryId: 1,
				Sender:  testutil.TestOwnerAddress,
				Result:  nil,
			},
			types.ErrEmptyResult,
		},
		{
			"empty kv results and block result",
			types.MsgSubmitQueryResult{
				QueryId: 1,
				Sender:  testutil.TestOwnerAddress,
				Result: &types.QueryResult{
					KvResults: nil,
					Block:     nil,
					Height:    100,
					Revision:  1,
				},
			},
			types.ErrEmptyResult,
		},
		{
			"zero query id",
			types.MsgSubmitQueryResult{
				QueryId: 0,
				Sender:  testutil.TestOwnerAddress,
				Result: &types.QueryResult{
					KvResults: []*types.StorageValue{{
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
			},
			types.ErrInvalidQueryID,
		},
		{
			"empty sender",
			types.MsgSubmitQueryResult{
				QueryId: 1,
				Sender:  "",
				Result: &types.QueryResult{
					KvResults: []*types.StorageValue{{
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
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid sender",
			types.MsgSubmitQueryResult{
				QueryId: 1,
				Sender:  "invalid_sender",
				Result: &types.QueryResult{
					KvResults: []*types.StorageValue{{
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
			},
			sdkerrors.ErrInvalidAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.SubmitQueryResult(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestMsgRemoveInterchainQueryRequestValidate(t *testing.T) {
	k, ctx := testkeeper.InterchainQueriesKeeper(t, nil, nil, nil, nil)
	msgServer := keeper.NewMsgServerImpl(*k)

	tests := []struct {
		name        string
		msg         types.MsgRemoveInterchainQueryRequest
		expectedErr error
	}{
		{
			"invalid query id",
			types.MsgRemoveInterchainQueryRequest{
				QueryId: 0,
				Sender:  testutil.TestOwnerAddress,
			},
			types.ErrInvalidQueryID,
		},
		{
			"empty sender",
			types.MsgRemoveInterchainQueryRequest{
				QueryId: 1,
				Sender:  "",
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid sender",
			types.MsgRemoveInterchainQueryRequest{
				QueryId: 1,
				Sender:  "invalid-sender",
			},
			sdkerrors.ErrInvalidAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.RemoveInterchainQuery(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestMsgUpdateInterchainQueryRequestValidate(t *testing.T) {
	k, ctx := testkeeper.InterchainQueriesKeeper(t, nil, nil, nil, nil)
	msgServer := keeper.NewMsgServerImpl(*k)

	tests := []struct {
		name        string
		msg         types.MsgUpdateInterchainQueryRequest
		expectedErr error
	}{
		{
			"invalid query id",
			types.MsgUpdateInterchainQueryRequest{
				QueryId: 0,
				NewKeys: []*types.KVKey{{
					Path: "staking",
					Key:  []byte{1, 2, 3},
				}},
				NewUpdatePeriod: 10,
				Sender:          testutil.TestOwnerAddress,
			},
			types.ErrInvalidQueryID,
		},
		{
			"empty keys, update_period and tx filter",
			types.MsgUpdateInterchainQueryRequest{
				QueryId:               1,
				NewKeys:               nil,
				NewUpdatePeriod:       0,
				NewTransactionsFilter: "",
				Sender:                testutil.TestOwnerAddress,
			},
			sdkerrors.ErrInvalidRequest,
		},
		{
			"both keys and filter sent",
			types.MsgUpdateInterchainQueryRequest{
				QueryId: 1,
				NewKeys: []*types.KVKey{{
					Path: "staking",
					Key:  []byte{1, 2, 3},
				}},
				NewUpdatePeriod:       0,
				NewTransactionsFilter: `{"field":"transfer.recipient","op":"eq","value":"cosmos1xxx"}`,
				Sender:                testutil.TestOwnerAddress,
			},
			sdkerrors.ErrInvalidRequest,
		},
		{
			"too many keys",
			types.MsgUpdateInterchainQueryRequest{
				QueryId:         1,
				NewKeys:         make([]*types.KVKey, types.DefaultMaxKvQueryKeysCount+1),
				NewUpdatePeriod: 0,
				Sender:          testutil.TestOwnerAddress,
			},
			types.ErrTooManyKVQueryKeys,
		},
		{
			"nil key",
			types.MsgUpdateInterchainQueryRequest{
				QueryId:         1,
				NewKeys:         []*types.KVKey{{Key: []byte("key1"), Path: "path1"}, nil},
				NewUpdatePeriod: 0,
				Sender:          testutil.TestOwnerAddress,
			},
			sdkerrors.ErrInvalidType,
		},
		{
			"duplicated keys",
			types.MsgUpdateInterchainQueryRequest{
				QueryId:         1,
				NewKeys:         []*types.KVKey{{Key: []byte("key1"), Path: "path1"}, {Key: []byte("key1"), Path: "path1"}},
				NewUpdatePeriod: 0,
				Sender:          testutil.TestOwnerAddress,
			},
			sdkerrors.ErrInvalidRequest,
		},
		{
			"empty key path",
			types.MsgUpdateInterchainQueryRequest{
				QueryId:         1,
				NewKeys:         []*types.KVKey{{Key: []byte("key1"), Path: ""}},
				NewUpdatePeriod: 0,
				Sender:          testutil.TestOwnerAddress,
			},
			types.ErrEmptyKeyPath,
		},
		{
			"empty key id",
			types.MsgUpdateInterchainQueryRequest{
				QueryId:         1,
				NewKeys:         []*types.KVKey{{Key: []byte(""), Path: "path"}},
				NewUpdatePeriod: 0,
				Sender:          testutil.TestOwnerAddress,
			},
			types.ErrEmptyKeyID,
		},
		{
			"empty sender",
			types.MsgUpdateInterchainQueryRequest{
				QueryId: 1,
				NewKeys: []*types.KVKey{{
					Path: "staking",
					Key:  []byte{1, 2, 3},
				}},
				NewUpdatePeriod: 10,
				Sender:          "",
			},
			sdkerrors.ErrInvalidAddress,
		},
		{
			"invalid sender",
			types.MsgUpdateInterchainQueryRequest{
				QueryId: 1,
				NewKeys: []*types.KVKey{{
					Path: "staking",
					Key:  []byte{1, 2, 3},
				}},
				NewUpdatePeriod: 10,
				Sender:          "invalid-sender",
			},
			sdkerrors.ErrInvalidAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := msgServer.UpdateInterchainQuery(ctx, &tt.msg)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}

func TestMsgUpdateParamsValidate(t *testing.T) {
	k, ctx := testkeeper.InterchainQueriesKeeper(t, nil, nil, nil, nil)

	tests := []struct {
		name        string
		msg         types.MsgUpdateParams
		expectedErr string
	}{
		{
			"empty authority",
			types.MsgUpdateParams{
				Authority: "",
			},
			"authority is invalid",
		},
		{
			"invalid authority",
			types.MsgUpdateParams{
				Authority: "invalid authority",
			},
			"authority is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := k.UpdateParams(ctx, &tt.msg)
			require.ErrorContains(t, err, tt.expectedErr)
			require.Nil(t, resp)
		})
	}
}
