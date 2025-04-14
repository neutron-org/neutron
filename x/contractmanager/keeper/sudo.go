package keeper

/*
Wasm contracts have the special entrypoint called sudo. The main purpose of the entrypoint is to be called from a trusted cosmos module, e.g. via a governance process.
We use the entrypoint to send back an ibc acknowledgement for an ibc transaction.
*/

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"github.com/neutron-org/neutron/v6/x/contractmanager/types"
)

func (k Keeper) HasContractInfo(ctx context.Context, contractAddress sdk.AccAddress) bool {
	return k.wasmKeeper.HasContractInfo(ctx, contractAddress)
}

func PrepareSudoCallbackMessage(request channeltypes.Packet, ack *channeltypes.Acknowledgement) ([]byte, error) {
	m := types.MessageSudoCallback{}
	if ack != nil && ack.GetError() == "" { //nolint:gocritic //
		m.Response = &types.ResponseSudoPayload{
			Data:    ack.GetResult(),
			Request: request,
		}
	} else if ack != nil {
		m.Error = &types.ErrorSudoPayload{
			Request: request,
			Details: ack.GetError(),
		}
	} else {
		m.Timeout = &types.TimeoutPayload{Request: request}
	}
	data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MessageSudoCallback: %v", err)
	}
	return data, nil
}

func PrepareOpenAckCallbackMessage(details types.OpenAckDetails) ([]byte, error) {
	x := types.MessageOnChanOpenAck{
		OpenAck: details,
	}
	m, err := json.Marshal(x)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MessageOnChanOpenAck: %v", err)
	}
	return m, nil
}

// SudoTxQueryResult is used to pass a tx query result to the contract that registered the query
// to:
//  1. check whether the transaction actually satisfies the initial query arguments;
//  2. execute business logic related to the tx query result / save the result to state.
func (k Keeper) SudoTxQueryResult(
	ctx context.Context,
	contractAddress sdk.AccAddress,
	queryID uint64,
	height ibcclienttypes.Height,
	data []byte,
) ([]byte, error) {
	c := sdk.UnwrapSDKContext(ctx)

	k.Logger(c).Debug("SudoTxQueryResult", "contractAddress", contractAddress)

	if !k.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		k.Logger(c).Debug("SudoTxQueryResult: contract not found", "contractAddress", contractAddress)
		return nil, fmt.Errorf("%s is not a contract address", contractAddress)
	}

	x := types.MessageTxQueryResult{}
	x.TxQueryResult.QueryID = queryID
	x.TxQueryResult.Height = height
	x.TxQueryResult.Data = data

	m, err := json.Marshal(x)
	if err != nil {
		k.Logger(c).Error("SudoTxQueryResult: failed to marshal MessageTxQueryResult message",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to marshal MessageTxQueryResult: %v", err)
	}

	resp, err := k.wasmKeeper.Sudo(ctx, contractAddress, m)
	if err != nil {
		k.Logger(c).Debug("SudoTxQueryResult: failed to sudo",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to sudo: %v", err)
	}

	return resp, nil
}

// SudoKVQueryResult is used to pass a kv query id to the contract that registered the query
// when a query result is provided by the relayer.
func (k Keeper) SudoKVQueryResult(
	ctx context.Context,
	contractAddress sdk.AccAddress,
	queryID uint64,
) ([]byte, error) {
	c := sdk.UnwrapSDKContext(ctx)

	k.Logger(c).Info("SudoKVQueryResult", "contractAddress", contractAddress)

	if !k.wasmKeeper.HasContractInfo(ctx, contractAddress) {
		k.Logger(c).Debug("SudoKVQueryResult: contract was not found", "contractAddress", contractAddress)
		return nil, fmt.Errorf("%s is not a contract address", contractAddress)
	}

	x := types.MessageKVQueryResult{}
	x.KVQueryResult.QueryID = queryID

	m, err := json.Marshal(x)
	if err != nil {
		k.Logger(c).Error("SudoKVQueryResult: failed to marshal MessageKVQueryResult message",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to marshal MessageKVQueryResult: %v", err)
	}

	resp, err := k.wasmKeeper.Sudo(ctx, contractAddress, m)
	if err != nil {
		k.Logger(c).Debug("SudoKVQueryResult: failed to sudo",
			"error", err, "contract_address", contractAddress)
		return nil, fmt.Errorf("failed to sudo: %v", err)
	}

	return resp, nil
}
