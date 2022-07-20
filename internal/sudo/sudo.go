package sudo

import (
	"encoding/json"
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/tendermint/tendermint/libs/log"
)

type SudoMessageTimeout struct {
	Timeout struct {
		Request channeltypes.Packet `json:"request"`
	} `json:"timeout"`
}
type SudoMessageResponse struct {
	Response struct {
		Request channeltypes.Packet `json:"request"`
		Message []byte              `json:"data"`
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
	x.Response.Message = msg
	x.Response.Request = request
	m, err := json.Marshal(x)
	if err != nil {
		s.Logger(ctx).Error("failed to marshal sudo message", "error", err, "request", request)
		return nil, sdkerrors.Wrap(err, "failed to marshal SudoMessage")
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
		s.Logger(ctx).Error("failed to marshal sudo message", "error", err, "request", request)
		return nil, sdkerrors.Wrap(err, "failed to marshal SudoMessage")
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
		s.Logger(ctx).Error("failed to marshal sudo message", "error", err, "request", request)
		return nil, sdkerrors.Wrap(err, "failed to marshal SudoMessage")
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
		s.Logger(ctx).Error("failed to marshal sudo message", "error", err)
		return nil, sdkerrors.Wrap(err, "failed to marshal SudoMessage")
	}
	s.Logger(ctx).Info("SudoOpenAck sending request", "data", string(m))

	r, err := s.wasmKeeper.Sudo(ctx, contractAddress, m)
	s.Logger(ctx).Info("SudoOpenAck received response", "err", err, "response", string(r))

	return r, err
}
