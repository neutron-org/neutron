package wasmbinding

import (
	"encoding/json"

	"cosmossdk.io/errors"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/wasmbinding/bindings"
)

// CustomQuerier returns a function that is an implementation of custom querier mechanism for specific messages
func CustomQuerier(qp *QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var contractQuery bindings.NeutronQuery
		if err := json.Unmarshal(request, &contractQuery); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal neutron query: %v", err)
		}

		switch {
		case contractQuery.InterchainQueryResult != nil:
			queryID := contractQuery.InterchainQueryResult.QueryID

			response, err := qp.GetInterchainQueryResult(ctx, queryID)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get interchain query result: %v", err)
			}

			bz, err := json.Marshal(response)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to marshal interchain query result: %v", err)
			}

			return bz, nil
		case contractQuery.InterchainAccountAddress != nil:

			interchainAccountAddress, err := qp.GetInterchainAccountAddress(ctx, contractQuery.InterchainAccountAddress)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get interchain account address: %v", err)
			}

			bz, err := json.Marshal(interchainAccountAddress)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to marshal interchain account query response: %v", err)
			}

			return bz, nil
		case contractQuery.RegisteredInterchainQueries != nil:
			registeredQueries, err := qp.GetRegisteredInterchainQueries(ctx, contractQuery.RegisteredInterchainQueries)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get registered queries: %v", err)
			}

			bz, err := json.Marshal(registeredQueries)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to marshal interchain account query response: %v", err)
			}

			return bz, nil
		case contractQuery.RegisteredInterchainQuery != nil:
			registeredQuery, err := qp.GetRegisteredInterchainQuery(ctx, contractQuery.RegisteredInterchainQuery)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get registered queries: %v", err)
			}

			bz, err := json.Marshal(registeredQuery)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to marshal interchain account query response: %v", err)
			}

			return bz, nil
		case contractQuery.TotalBurnedNeutronsAmount != nil:
			totalBurnedNeutrons, err := qp.GetTotalBurnedNeutronsAmount(ctx, contractQuery.TotalBurnedNeutronsAmount)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get total burned neutrons amount: %v", err)
			}

			bz, err := json.Marshal(totalBurnedNeutrons)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to marshal total burned neutrons amount response: %v", err)
			}

			return bz, nil
		case contractQuery.MinIbcFee != nil:
			minFee, err := qp.GetMinIbcFee(ctx, contractQuery.MinIbcFee)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get min fee: %v", err)
			}

			bz, err := json.Marshal(minFee)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to marshal min fee response: %v", err)
			}

			return bz, nil

		case contractQuery.FullDenom != nil:
			creator := contractQuery.FullDenom.CreatorAddr
			subdenom := contractQuery.FullDenom.Subdenom

			fullDenom, err := GetFullDenom(creator, subdenom)
			if err != nil {
				return nil, errors.Wrap(err, "unable to get full denom")
			}

			res := bindings.FullDenomResponse{
				Denom: fullDenom,
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, errors.Wrap(err, "failed to JSON marshal FullDenomResponse response")
			}

			return bz, nil

		case contractQuery.DenomAdmin != nil:
			res, err := qp.GetDenomAdmin(ctx, contractQuery.DenomAdmin.Subdenom)
			if err != nil {
				return nil, errors.Wrap(err, "unable to get denom admin")
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, errors.Wrap(err, "failed to JSON marshal DenomAdminResponse response")
			}

			return bz, nil

		case contractQuery.BeforeSendHook != nil:
			res, err := qp.GetBeforeSendHook(ctx, contractQuery.BeforeSendHook.Denom)
			if err != nil {
				return nil, errors.Wrap(err, "unable to get denom before_send_hook")
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, errors.Wrap(err, "failed to JSON marshal BeforeSendHookResponse response")
			}

			return bz, nil

		case contractQuery.Failures != nil:
			res, err := qp.GetFailures(ctx, contractQuery.Failures.Address, contractQuery.Failures.Pagination)
			if err != nil {
				return nil, errors.Wrap(err, "unable to get denom admin")
			}

			bz, err := json.Marshal(res)
			if err != nil {
				return nil, errors.Wrap(err, "failed to JSON marshal FailuresResponse response")
			}

			return bz, nil

		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown neutron query type"}
		}
	}
}
