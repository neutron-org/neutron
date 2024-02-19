package types

import (
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

// MessageTxQueryResult is passed to a contract's sudo() entrypoint when a tx was submitted
// for a transaction query.
type MessageTxQueryResult struct {
	TxQueryResult struct {
		QueryID uint64                `json:"query_id"`
		Height  ibcclienttypes.Height `json:"height"`
		Data    []byte                `json:"data"`
	} `json:"tx_query_result"`
}

// MessageKVQueryResult is passed to a contract's sudo() entrypoint when a result
// was submitted for a kv-query.
type MessageKVQueryResult struct {
	KVQueryResult struct {
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
