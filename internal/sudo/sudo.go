package sudo

/*
Wasm contracts have the special entrypoint called sudo. The main purpose of the entrypoint is to be called from a trusted cosmos module, e.g. via a governance process.
We use the entrypoint to send back an ibc acknowledgement for an ibc transaction.
The package contains the code to postprocess incoming from a relayer acknowledgement and pass it to the  ibc transaction contract initiator
*/

import (
	"encoding/json"
	"fmt"

	"github.com/CosmWasm/wasmd/x/wasm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/tendermint/tendermint/libs/log"
)

// MessageTxQueryResult is passed to a contract's sudo() entrypoint when a tx was submitted
// for a transaction query.
type MessageTxQueryResult struct {
	TxQueryResult struct {
		QueryID uint64 `json:"query_id"`
		Height  uint64 `json:"height"`
		Data    []byte `json:"data"`
	} `json:"tx_query_result"`
}

// MessageKVQueryResult is passed to a contract's sudo() entrypoint when a result
// was submitted for a kv-query.
type MessageKVQueryResult struct {
	KVQueryResult struct {
		QueryID uint64 `json:"query_id"`
	} `json:"kv_query_result"`
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
	CounterpartyChannelId string `json:"counterparty_channel_id"`
	CounterpartyVersion   string `json:"counterparty_version"`
}

type Handler struct {
	moduleName string
	wasmKeeper *wasm.Keeper
}

func (s *Handler) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", s.moduleName))
}

func NewSudoHandler(wasmKeeper *wasm.Keeper, moduleName string) Handler {
	return Handler{
		moduleName: moduleName,
		wasmKeeper: wasmKeeper,
	}
}

func (s *Handler) SudoResponse(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	request channeltypes.Packet,
	msg []byte,
) ([]byte, error) {
	s.Logger(ctx).Debug("SudoResponse", "contractAddress", contractAddress, "request", request, "msg", msg)

	if !s.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		s.Logger(ctx).Debug("SudoResponse: contract not found", "contractAddress", contractAddress)
		return nil, fmt.Errorf("%s is not a contract address", contractAddress)
	}

	x := MessageResponse{}
	x.Response.Data = msg
	x.Response.Request = request
	m, err := json.Marshal(x)
	if err != nil {
		s.Logger(ctx).Error("SudoResponse: failed to marshal MessageResponse message",
			"error", err, "request", request, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to marshal MessageResponse: %v", err)
	}

	resp, err := s.wasmKeeper.Sudo(ctx, contractAddress, m)
	if err != nil {
		s.Logger(ctx).Debug("SudoResponse: failed to Sudo",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to Sudo: %v", err)
	}

	return resp, nil
}

func (s *Handler) SudoTimeout(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	request channeltypes.Packet,
) ([]byte, error) {
	s.Logger(ctx).Info("SudoTimeout", "contractAddress", contractAddress, "request", request)

	if !s.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		s.Logger(ctx).Debug("SudoTimeout: contract not found", "contractAddress", contractAddress)
		return nil, fmt.Errorf("%s is not a contract address", contractAddress)
	}

	x := MessageTimeout{}
	x.Timeout.Request = request
	m, err := json.Marshal(x)
	if err != nil {
		s.Logger(ctx).Error("failed to marshal MessageTimeout message",
			"error", err, "request", request, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to marshal MessageTimeout: %v", err)
	}

	s.Logger(ctx).Info("SudoTimeout sending request", "data", string(m))

	resp, err := s.wasmKeeper.Sudo(ctx, contractAddress, m)
	if err != nil {
		s.Logger(ctx).Debug("SudoTimeout: failed to Sudo",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to Sudo: %v", err)
	}

	return resp, nil
}

func (s *Handler) SudoError(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	request channeltypes.Packet,
	details string,
) ([]byte, error) {
	s.Logger(ctx).Debug("SudoError", "contractAddress", contractAddress, "request", request)

	if !s.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		s.Logger(ctx).Debug("SudoError: contract not found", "contractAddress", contractAddress)
		return nil, fmt.Errorf("%s is not a contract address", contractAddress)
	}

	x := MessageError{}
	x.Error.Request = request
	x.Error.Details = details
	m, err := json.Marshal(x)
	if err != nil {
		s.Logger(ctx).Error("SudoError: failed to marshal MessageError message",
			"error", err, "contract_address", contractAddress, "request", request)
		return nil, fmt.Errorf("failed to marshal MessageError: %v", err)
	}

	resp, err := s.wasmKeeper.Sudo(ctx, contractAddress, m)
	if err != nil {
		s.Logger(ctx).Debug("SudoError: failed to Sudo",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to Sudo: %v", err)
	}

	return resp, nil
}

func (s *Handler) SudoOnChanOpenAck(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	details OpenAckDetails,
) ([]byte, error) {
	s.Logger(ctx).Debug("SudoOnChanOpenAck", "contractAddress", contractAddress)

	if !s.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		s.Logger(ctx).Debug("SudoOnChanOpenAck: contract not found", "contractAddress", contractAddress)
		return nil, fmt.Errorf("%s is not a contract address", contractAddress)
	}

	x := MessageOnChanOpenAck{}
	x.OpenAck = details
	m, err := json.Marshal(x)
	if err != nil {
		s.Logger(ctx).Error("SudoOnChanOpenAck: failed to marshal MessageOnChanOpenAck message",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to marshal MessageOnChanOpenAck: %v", err)
	}
	s.Logger(ctx).Info("SudoOnChanOpenAck sending request", "data", string(m))

	resp, err := s.wasmKeeper.Sudo(ctx, contractAddress, m)
	if err != nil {
		s.Logger(ctx).Debug("SudoOnChanOpenAck: failed to Sudo",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to Sudo: %v", err)
	}

	return resp, nil
}

// SudoTxQueryResult is used to pass a tx query result to the contract that registered the query
// to:
// 		1. check whether the transaction actually satisfies the initial query arguments;
// 		2. execute business logic related to the tx query result / save the result to state.
func (s *Handler) SudoTxQueryResult(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	queryID uint64,
	height int64,
	data []byte,
) ([]byte, error) {
	s.Logger(ctx).Debug("SudoTxQueryResult", "contractAddress", contractAddress)

	if !s.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		s.Logger(ctx).Debug("SudoTxQueryResult: contract not found", "contractAddress", contractAddress)
		return nil, fmt.Errorf("%s is not a contract address", contractAddress)
	}

	x := MessageTxQueryResult{}
	x.TxQueryResult.QueryID = queryID
	x.TxQueryResult.Height = uint64(height)
	x.TxQueryResult.Data = data

	m, err := json.Marshal(x)
	if err != nil {
		s.Logger(ctx).Error("SudoTxQueryResult: failed to marshal MessageTxQueryResult message",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to marshal MessageTxQueryResult: %v", err)
	}

	resp, err := s.wasmKeeper.Sudo(ctx, contractAddress, m)
	if err != nil {
		s.Logger(ctx).Debug("SudoTxQueryResult: failed to Sudo",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to Sudo: %v", err)
	}

	return resp, nil
}

// SudoKVQueryResult is used to pass a kv query id to the contract that registered the query
// when a query result is provided by the relayer.
func (s *Handler) SudoKVQueryResult(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	queryID uint64,
) ([]byte, error) {
	s.Logger(ctx).Info("SudoKVQueryResult", "contractAddress", contractAddress)

	if !s.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		s.Logger(ctx).Debug("SudoKVQueryResult: contract was not found", "contractAddress", contractAddress)
		return nil, fmt.Errorf("%s is not a contract address", contractAddress)
	}

	x := MessageKVQueryResult{}
	x.KVQueryResult.QueryID = queryID

	m, err := json.Marshal(x)
	if err != nil {
		s.Logger(ctx).Error("SudoKVQueryResult: failed to marshal MessageKVQueryResult message",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to marshal MessageKVQueryResult: %v", err)
	}

	resp, err := s.wasmKeeper.Sudo(ctx, contractAddress, m)
	if err != nil {
		s.Logger(ctx).Debug("SudoKVQueryResult: failed to Sudo",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to Sudo: %v", err)
	}

	return resp, nil
}
