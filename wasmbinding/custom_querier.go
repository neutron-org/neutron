package wasmbinding

import (
	"encoding/json"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/neutron-org/neutron/wasmbinding/bindings"
)

// CustomQuerier returns a function that is an implementation of custom querier mechanism for specific messages
func CustomQuerier(qp *QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var contractQuery bindings.NeutronQuery
		if err := json.Unmarshal(request, &contractQuery); err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to unmarshal neutron query: %v", err)
		}

		switch {
		case contractQuery.InterchainQueryResult != nil:
			queryID := contractQuery.InterchainQueryResult.QueryID

			response, err := qp.GetInterchainQueryResult(ctx, queryID)
			if err != nil {
				return nil, sdkerrors.Wrapf(err, "failed to get interchain query result: %v", err)
			}

			res := bindings.InterchainQueryResultResponse{
				Result: response,
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, sdkerrors.Wrapf(err, "failed to marshal interchain query result: %v", err)
			}

			return bz, nil
		case contractQuery.InterchainAccountAddress != nil:
			ownerAddress := contractQuery.InterchainAccountAddress.OwnerAddress
			connectionID := contractQuery.InterchainAccountAddress.ConnectionID

			interchainAccountAddress, err := qp.GetInterchainAccountAddress(ctx, ownerAddress, connectionID)
			if err != nil {
				return nil, sdkerrors.Wrapf(err, "failed to get interchain account address: %v", err)
			}

			res := bindings.InterchainAccountAddressResponse{
				InterchainAccountAddress: &interchainAccountAddress,
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, sdkerrors.Wrapf(err, "failed to marshal interchain account query response: %v", err)
			}

			return bz, nil
		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown neutron query type"}
		}
	}
}
