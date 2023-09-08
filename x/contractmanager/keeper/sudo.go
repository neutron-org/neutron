package keeper

/*
Wasm contracts have the special entrypoint called sudo. The main purpose of the entrypoint is to be called from a trusted cosmos module, e.g. via a governance process.
We use the entrypoint to send back an ibc acknowledgement for an ibc transaction.
*/

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/neutron-org/neutron/x/contractmanager/types"
)

func (k Keeper) HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool {
	return k.wasmKeeper.HasContractInfo(ctx, contractAddress)
}

func prepareSudoCallbackMessage(request channeltypes.Packet, ack *channeltypes.Acknowledgement) types.MessageSudoCallback {
	m := types.MessageSudoCallback{
		Response: nil,
		Error:    nil,
		Timeout:  nil,
	}
	return m
}

func (k Keeper) SudoResponse(
	ctx sdk.Context,
	senderAddress sdk.AccAddress,
	request channeltypes.Packet,
	ack channeltypes.Acknowledgement,
) ([]byte, error) {
	k.Logger(ctx).Debug("SudoResponse", "senderAddress", senderAddress, "request", request, "msg", ack.GetResult())

	x := types.MessageResponse{}
	x.Response.Data = ack.GetResult()
	x.Response.Request = request
	m, err := json.Marshal(x)
	if err != nil {
		k.Logger(ctx).Error("SudoResponse: failed to marshal MessageResponse message",
			"error", err, "request", request, "contract_address", senderAddress)
		return nil, fmt.Errorf("failed to marshal MessageResponse: %v", err)
	}

	resp, err := k.wasmKeeper.Sudo(ctx, senderAddress, m)
	if err != nil {
		k.Logger(ctx).Debug("SudoResponse: failed to sudo",
			"error", err, "contract_address", senderAddress)
		return nil, fmt.Errorf("failed to sudo: %v", err)
	}

	return resp, nil
}

func (k Keeper) SudoTimeout(
	ctx sdk.Context,
	senderAddress sdk.AccAddress,
	request channeltypes.Packet,
) ([]byte, error) {
	k.Logger(ctx).Info("SudoTimeout", "senderAddress", senderAddress, "request", request)

	x := types.MessageTimeout{}
	x.Timeout.Request = request
	m, err := json.Marshal(x)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal MessageTimeout message",
			"error", err, "request", request, "contract_address", senderAddress)
		return nil, fmt.Errorf("failed to marshal MessageTimeout: %v", err)
	}

	k.Logger(ctx).Info("SudoTimeout sending request", "data", string(m))

	resp, err := k.wasmKeeper.Sudo(ctx, senderAddress, m)
	if err != nil {
		k.Logger(ctx).Debug("SudoTimeout: failed to sudo",
			"error", err, "contract_address", senderAddress)
		return nil, fmt.Errorf("failed to sudo: %v", err)
	}

	return resp, nil
}

func (k Keeper) SudoError(
	ctx sdk.Context,
	senderAddress sdk.AccAddress,
	request channeltypes.Packet,
	ack channeltypes.Acknowledgement,
) ([]byte, error) {
	k.Logger(ctx).Debug("SudoError", "senderAddress", senderAddress, "request", request)

	x := types.MessageError{}
	x.Error.Request = request
	x.Error.Details = ack.GetError()
	m, err := json.Marshal(x)
	if err != nil {
		k.Logger(ctx).Error("SudoError: failed to marshal MessageError message",
			"error", err, "contract_address", senderAddress, "request", request)
		return nil, fmt.Errorf("failed to marshal MessageError: %v", err)
	}

	resp, err := k.wasmKeeper.Sudo(ctx, senderAddress, m)
	if err != nil {
		k.Logger(ctx).Debug("SudoError: failed to sudo",
			"error", err, "contract_address", senderAddress)
		return nil, fmt.Errorf("failed to sudo: %v", err)
	}

	return resp, nil
}

func (k Keeper) SudoOnChanOpenAck(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	details types.OpenAckDetails,
) ([]byte, error) {
	k.Logger(ctx).Debug("SudoOnChanOpenAck", "contractAddress", contractAddress)

	x := types.MessageOnChanOpenAck{}
	x.OpenAck = details
	m, err := json.Marshal(x)
	if err != nil {
		k.Logger(ctx).Error("SudoOnChanOpenAck: failed to marshal MessageOnChanOpenAck message",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to marshal MessageOnChanOpenAck: %v", err)
	}
	k.Logger(ctx).Info("SudoOnChanOpenAck sending request", "data", string(m))

	resp, err := k.wasmKeeper.Sudo(ctx, contractAddress, m)
	if err != nil {
		k.Logger(ctx).Debug("SudoOnChanOpenAck: failed to sudo",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to sudo: %v", err)
	}

	return resp, nil
}

// SudoTxQueryResult is used to pass a tx query result to the contract that registered the query
// to:
//  1. check whether the transaction actually satisfies the initial query arguments;
//  2. execute business logic related to the tx query result / save the result to state.
func (k Keeper) SudoTxQueryResult(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	queryID uint64,
	height ibcclienttypes.Height,
	data []byte,
) ([]byte, error) {
	k.Logger(ctx).Debug("SudoTxQueryResult", "contractAddress", contractAddress)

	if !k.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		k.Logger(ctx).Debug("SudoTxQueryResult: contract not found", "contractAddress", contractAddress)
		return nil, fmt.Errorf("%s is not a contract address", contractAddress)
	}

	x := types.MessageTxQueryResult{}
	x.TxQueryResult.QueryID = queryID
	x.TxQueryResult.Height = height
	x.TxQueryResult.Data = data

	m, err := json.Marshal(x)
	if err != nil {
		k.Logger(ctx).Error("SudoTxQueryResult: failed to marshal MessageTxQueryResult message",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to marshal MessageTxQueryResult: %v", err)
	}

	resp, err := k.wasmKeeper.Sudo(ctx, contractAddress, m)
	if err != nil {
		k.Logger(ctx).Debug("SudoTxQueryResult: failed to sudo",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to sudo: %v", err)
	}

	return resp, nil
}

// SudoKVQueryResult is used to pass a kv query id to the contract that registered the query
// when a query result is provided by the relayer.
func (k Keeper) SudoKVQueryResult(
	ctx sdk.Context,
	contractAddress sdk.AccAddress,
	queryID uint64,
) ([]byte, error) {
	k.Logger(ctx).Info("SudoKVQueryResult", "contractAddress", contractAddress)

	if !k.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		k.Logger(ctx).Debug("SudoKVQueryResult: contract was not found", "contractAddress", contractAddress)
		return nil, fmt.Errorf("%s is not a contract address", contractAddress)
	}

	x := types.MessageKVQueryResult{}
	x.KVQueryResult.QueryID = queryID

	m, err := json.Marshal(x)
	if err != nil {
		k.Logger(ctx).Error("SudoKVQueryResult: failed to marshal MessageKVQueryResult message",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to marshal MessageKVQueryResult: %v", err)
	}

	resp, err := k.wasmKeeper.Sudo(ctx, contractAddress, m)
	if err != nil {
		k.Logger(ctx).Debug("SudoKVQueryResult: failed to sudo",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to sudo: %v", err)
	}

	return resp, nil
}
