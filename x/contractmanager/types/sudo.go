package types

import (
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
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

type MessageSudoCallback struct {
	Response *ResponseSudoPayload `json:"response"`
	Error    *ErrorSudoPayload    `json:"error"`
	Timeout  *TimeoutPayload      `json:"timeout"`
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

// MessageTimeout is passed to a contract's sudo() entrypoint when an interchain
// transaction failed with a timeout.
type MessageTimeout struct {
	Timeout struct {
		Request channeltypes.Packet `json:"request"`
	} `json:"timeout"`
}

// MessageResponse is passed to a contract's sudo() entrypoint when an interchain
// transaction was executed successfully.
type MessageResponse struct {
	Response struct {
		Request channeltypes.Packet `json:"request"`
		Data    []byte              `json:"data"` // Message data
	} `json:"response"`
}

// MessageError is passed to a contract's sudo() entrypoint when an interchain
// transaction was executed with an error.
type MessageError struct {
	Error struct {
		Request channeltypes.Packet `json:"request"`
		Details string              `json:"details"`
	} `json:"error"`
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
