package types

import (
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

// MessageTxQueryResult is the model of the `sudo` message sent to a smart contract when a TX
// Interchain Query result is submitted. The owner of a TX Interchain Query must implement a `sudo`
// entry point to handle `tx_query_result` messages and include the necessary logic in it. The
// `tx_query_result` handler functions as a callback, triggered by the `interchainqueries` module
// each time a TX query result is submitted.
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

// MessageKvQueryResult is the model of the `sudo` message sent to a smart contract when a KV
// Interchain Query result is submitted. If the owner of a KV Interchain Query wants to handle
// query updates as part of the result submission transaction, they must implement a `sudo` entry
// point to process `kv_query_result` messages and include the necessary logic in it. The
// `kv_query_result` handler acts as a callback, triggered by the interchainqueries module whenever
// a KV query result is submitted.
//
// Note that the message does not include the actual query result, only the query ID. To access the
// result data, use the `Query/QueryResult` RPC of the `interchainqueries` module.
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
