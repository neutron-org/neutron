package types

import (
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

// MessageTxQueryResult is the model of the sudo message sent to a smart contract on a TX-typed
// Interchain Query result submission. The owner of a TX-typed Interchain Query must define a
// `sudo` entry_point for handling `tx_query_result` messages and place the needed logic there.
// The `tx_query_result` handler is treated by the interchainqueries module as a callback that is
// called each time a TX-typed query result is submitted.
type MessageTxQueryResult struct {
	TxQueryResult struct {
		// QueryID is the ID of the TX query which result is being submitted.
		QueryID uint64 `json:"query_id"`
		// Height is the remote chain's block height the transaction was included in.
		Height ibcclienttypes.Height `json:"height"`
		// Data is the body of the transaction.
		Data []byte `json:"data"`
	} `json:"tx_query_result"`
}

// MessageKVQueryResult is the model of the sudo message sent to a smart contract on a KV-typed
// Interchain Query result submission. If the owner of a KV-typed Interchain Query wants to handle
// the query updates, it must define a `sudo` entry_point for handling `kv_query_result` messages
// and place the needed logic there. The `kv_query_result` handler is treated by the
// interchainqueries module as a callback that is called each time a KV-typed query result is
// submitted.
//
// Note that there is no query result sent, only the query ID. In order to access the actual
// result, use the Query/QueryResult RPC of the interchainqueries module.
type MessageKVQueryResult struct {
	KVQueryResult struct {
		// QueryID is the ID of the KV query which result is being submitted.
		QueryID uint64 `json:"query_id"`
	} `json:"kv_query_result"`
}

// MessageSudoCallback is passed to a contract's sudo() entrypoint when an interchain
// transaction ended up with Success/Error or timed out.
type MessageSudoCallback struct {
	Response *ResponseSudoPayload `json:"response,omitempty"`
	Error    *ErrorSudoPayload    `json:"error,omitempty"`
	Timeout  *TimeoutPayload      `json:"timeout,omitempty"`
}

type ResponseSudoPayload struct {
	Request channeltypes.Packet `json:"request"`
	Data    []byte              `json:"data"` // Message data
}

type ErrorSudoPayload struct {
	Request channeltypes.Packet `json:"request"`
	Details string              `json:"details"`
}

type TimeoutPayload struct {
	Request channeltypes.Packet `json:"request"`
}

// MessageOnChanOpenAck is passed to a contract's sudo() entrypoint when an interchain
// account was successfully  registered.
type MessageOnChanOpenAck struct {
	OpenAck OpenAckDetails `json:"open_ack"`
}

type OpenAckDetails struct {
	PortID                string `json:"port_id"`
	ChannelID             string `json:"channel_id"`
	CounterpartyChannelID string `json:"counterparty_channel_id"`
	CounterpartyVersion   string `json:"counterparty_version"`
}
