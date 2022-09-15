//nolint:revive,stylecheck  // if we change the names of var-naming things here, we harm some kind of mapping.
package bindings

import "github.com/neutron-org/neutron/x/interchainqueries/types"

// ProtobufAny is a hack-struct to serialize protobuf Any message into JSON object
type ProtobufAny struct {
	TypeURL string `json:"type_url"`
	Value   []byte `json:"value"`
}

// NeutronMsg is used like a sum type to hold one of custom Neutron messages.
// Follow https://github.com/neutron-org/neutron-contracts/tree/main/packages/bindings/src/msg.rs
// for more information.
type NeutronMsg struct {
	SubmitTx                  *SubmitTx                  `json:"submit_tx,omitempty"`
	RegisterInterchainAccount *RegisterInterchainAccount `json:"register_interchain_account,omitempty"`
	RegisterInterchainQuery   *RegisterInterchainQuery   `json:"register_interchain_query,omitempty"`
	UpdateInterchainQuery     *UpdateInterchainQuery     `json:"update_interchain_query,omitempty"`
	RemoveInterchainQuery     *RemoveInterchainQuery     `json:"remove_interchain_query,omitempty"`
	AddAdmin                  *AddAdmin                  `json:"add_admin,omitempty"`
}

// SubmitTx submits interchain transaction on a remote chain.
type SubmitTx struct {
	ConnectionId        string        `json:"connection_id"`
	InterchainAccountId string        `json:"interchain_account_id"`
	Msgs                []ProtobufAny `json:"msgs"`
	Memo                string        `json:"memo"`
	Timeout             uint64        `json:"timeout"`
}

// SubmitTxResponse holds response from SubmitTx.
type SubmitTxResponse struct {
	// SequenceId is a channel's sequence_id for outgoing ibc packet. Unique per a channel.
	SequenceId uint64 `json:"sequence_id"`
	// Channel is a src channel on neutron side transaction was submitted from
	Channel string `json:"channel"`
}

// RegisterInterchainAccount creates account on remote chain.
type RegisterInterchainAccount struct {
	ConnectionId        string `json:"connection_id"`
	InterchainAccountId string `json:"interchain_account_id"`
}

// RegisterInterchainAccountResponse holds response for RegisterInterchainAccount.
type RegisterInterchainAccountResponse struct{}

// RegisterInterchainQuery creates a query for remote chain.
type RegisterInterchainQuery struct {
	QueryType          string         `json:"query_type"`
	Keys               []*types.KVKey `json:"keys"`
	TransactionsFilter string         `json:"transactions_filter"`
	ConnectionId       string         `json:"connection_id"`
	UpdatePeriod       uint64         `json:"update_period"`
}

type AddAdmin struct {
	Admin string `protobuf:"bytes,2,opt,name=admin,proto3" json:"admin,omitempty"`
}

type AddAdminResponse struct {
}

// RegisterInterchainQueryResponse holds response for RegisterInterchainQuery
type RegisterInterchainQueryResponse struct {
	Id uint64 `json:"id"`
}

type RemoveInterchainQuery struct {
	QueryId uint64 `json:"query_id"`
}

type RemoveInterchainQueryResponse struct{}

type UpdateInterchainQuery struct {
	QueryId         uint64         `json:"query_id,omitempty"`
	NewKeys         []*types.KVKey `json:"new_keys,omitempty"`
	NewUpdatePeriod uint64         `json:"new_update_period,omitempty"`
}

type UpdateInterchainQueryResponse struct{}
