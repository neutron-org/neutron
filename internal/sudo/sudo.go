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
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/tendermint/tendermint/libs/log"
)

type SudoMessageCheckTxQueryResult struct {
	CheckTxQueryResult struct {
		QueryID uint64 `json:"query_id"`
		Height  uint64 `json:"height"`
		Data    []byte `json:"data"`
	} `json:"check_tx_query_result"`
}

type SudoMessageTimeout struct {
	Timeout struct {
		Request channeltypes.Packet `json:"request"`
	} `json:"timeout"`
}

type SudoMessageResponse struct {
	Response struct {
		Request channeltypes.Packet `json:"request"`
		Data    []byte              `json:"data"` // Message data
	} `json:"response"`
}

type SudoMessageError struct {
	Error struct {
		Request channeltypes.Packet `json:"request"`
		Details string              `json:"details"`
	} `json:"error"`
}

type SudoMessageOpenAck struct {
	OpenAck OpenAckDetails `json:"open_ack"`
}

type OpenAckDetails struct {
	PortID                string `json:"port_id"`
	ChannelID             string `json:"channel_id"`
	CounterpartyChannelId string `json:"counterparty_channel_id"`
	CounterpartyVersion   string `json:"counterparty_version"`
}

type SudoHandler struct {
	moduleName string
	wasmKeeper *wasm.Keeper
}

func (s *SudoHandler) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", s.moduleName))
}

func NewSudoHandler(wasmKeeper *wasm.Keeper, moduleName string) SudoHandler {
	return SudoHandler{
		moduleName: moduleName,
		wasmKeeper: wasmKeeper,
	}
}

func (s *SudoHandler) SudoResponse(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	request channeltypes.Packet,
	msg []byte,
) ([]byte, error) {
	s.Logger(ctx).Info("SudoResponse", "contractAddress", contractAddress, "request", request, "msg", msg)

	// TODO: basically just for unit tests right now. But i think we will have the same logic in the production
	if !s.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		return nil, nil
	}

	x := SudoMessageResponse{}
	x.Response.Data = msg
	x.Response.Request = request
	m, err := json.Marshal(x)
	if err != nil {
		s.Logger(ctx).Error("failed to marshal SudoMessageResponse message", "error", err, "request", request)
		return nil, sdkerrors.Wrap(err, "failed to marshal SudoMessageResponse message")
	}

	s.Logger(ctx).Info("SudoResponse sending request", "data", string(m))

	r, err := s.wasmKeeper.Sudo(ctx, contractAddress, m)
	s.Logger(ctx).Info("SudoResponse received response", "err", err, "response", string(r))

	return r, err
}

func (s *SudoHandler) SudoTimeout(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	request channeltypes.Packet,
) ([]byte, error) {
	s.Logger(ctx).Info("SudoTimeout", "contractAddress", contractAddress, "request", request)

	// TODO: basically just for unit tests right now. But i think we will have the same logic in the production
	if !s.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		return nil, nil
	}

	x := SudoMessageTimeout{}
	x.Timeout.Request = request
	m, err := json.Marshal(x)
	if err != nil {
		s.Logger(ctx).Error("failed to marshal SudoMessageTimeout message", "error", err, "request", request)
		return nil, sdkerrors.Wrap(err, "failed to marshal SudoMessageTimeout message")
	}

	s.Logger(ctx).Info("SudoTimeout sending request", "data", string(m))

	r, err := s.wasmKeeper.Sudo(ctx, contractAddress, m)
	s.Logger(ctx).Info("SudoTimeout received response", "err", err, "response", string(r))

	return r, err
}

func (s *SudoHandler) SudoError(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	request channeltypes.Packet,
	details string,
) ([]byte, error) {
	s.Logger(ctx).Info("SudoError", "contractAddress", contractAddress, "request", request)

	// TODO: basically just for unit tests right now. But i think we will have the same logic in the production
	if !s.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		return nil, nil
	}

	x := SudoMessageError{}
	x.Error.Request = request
	x.Error.Details = details
	m, err := json.Marshal(x)
	if err != nil {
		s.Logger(ctx).Error("failed to marshal SudoMessageError message", "error", err, "request", request)
		return nil, sdkerrors.Wrap(err, "failed to marshal SudoMessageError message")
	}

	s.Logger(ctx).Info("SudoError sending request", "data", string(m))

	r, err := s.wasmKeeper.Sudo(ctx, contractAddress, m)
	s.Logger(ctx).Info("SudoError received response", "err", err, "response", string(r))

	return r, err
}

func (s *SudoHandler) SudoOpenAck(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	details OpenAckDetails,
) ([]byte, error) {
	s.Logger(ctx).Info("SudoOpenAck", "contractAddress", contractAddress)

	// TODO: basically just for unit tests right now. But i think we will have the same logic in the production
	if !s.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		return nil, nil
	}

	x := SudoMessageOpenAck{}
	x.OpenAck = details
	m, err := json.Marshal(x)
	if err != nil {
		s.Logger(ctx).Error("failed to marshal SudoMessageOpenAck message", "error", err)
		return nil, sdkerrors.Wrap(err, "failed to marshal SudoMessageOpenAck message")
	}
	s.Logger(ctx).Info("SudoOpenAck sending request", "data", string(m))

	r, err := s.wasmKeeper.Sudo(ctx, contractAddress, m)
	s.Logger(ctx).Info("SudoOpenAck received response", "err", err, "response", string(r))

	return r, err
}

// SudoCheckTxQueryResult is used to pass a tx query result to the contract that registered the query
// to check whether the transaction actually satisfies the initial query arguments.
//
// Please note that this callback can be potentially used by the contact to execute business logic.
func (s *SudoHandler) SudoCheckTxQueryResult(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	height int64,
	data []byte,
) ([]byte, error) {
	s.Logger(ctx).Info("SudoCheckTxQueryResult", "contractAddress", contractAddress)

	// TODO: basically just for unit tests right now. But i think we will have the same logic in the production
	if !s.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		return nil, nil
	}

	x := SudoMessageCheckTxQueryResult{}
	x.CheckTxQueryResult.Height = uint64(height)
	x.CheckTxQueryResult.Data = data

	m, err := json.Marshal(x)
	if err != nil {
		s.Logger(ctx).Error("failed to marshal SudoMessageCheckTxQueryResult message", "error", err)
		return nil, sdkerrors.Wrap(err, "failed to marshal SudoMessageCheckTxQueryResult message")
	}

	r, err := s.wasmKeeper.Sudo(ctx, contractAddress, m)
	s.Logger(ctx).Info("SudoOpenAck received response", "err", err, "response", string(r))

	return r, err
}
