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

		case contractQuery.Incentives != nil:
			incentives := contractQuery.Incentives
			switch {
			case incentives.ModuleStatus != nil:
				res, err := qp.GetModuleStatus(ctx)
				if err != nil {
					return nil, errors.Wrap(err, "unable to query ModuleStatus")
				}

				bz, err := json.Marshal(res)
				if err != nil {
					return nil, errors.Wrap(err, "failed to JSON marshal ModuleStatus response")
				}

				return bz, nil

			case incentives.GaugeByID != nil:
				res, err := qp.GetGaugeByID(ctx, incentives.GaugeByID.ID)
				if err != nil {
					return nil, errors.Wrap(err, "unable to query GaugeByID")
				}

				bz, err := json.Marshal(res)
				if err != nil {
					return nil, errors.Wrap(err, "failed to JSON marshal GaugeByID response")
				}

				return bz, nil

			case incentives.Gauges != nil:
				res, err := qp.GetGauges(ctx, incentives.Gauges.Status, incentives.Gauges.Denom)
				if err != nil {
					return nil, errors.Wrap(err, "unable to query Gauges")
				}

				bz, err := json.Marshal(res)
				if err != nil {
					return nil, errors.Wrap(err, "failed to JSON marshal Gauges response")
				}

				return bz, nil

			case incentives.StakeByID != nil:
				res, err := qp.GetStakeByID(ctx, incentives.StakeByID.StakeID)
				if err != nil {
					return nil, errors.Wrap(err, "unable to query StakeByID")
				}

				bz, err := json.Marshal(res)
				if err != nil {
					return nil, errors.Wrap(err, "failed to JSON marshal StakeByID response")
				}

				return bz, nil

			case incentives.Stakes != nil:
				res, err := qp.GetStakes(ctx, incentives.Stakes.Owner)
				if err != nil {
					return nil, errors.Wrap(err, "unable to query Stakes")
				}

				bz, err := json.Marshal(res)
				if err != nil {
					return nil, errors.Wrap(err, "failed to JSON marshal Stakes response")
				}

				return bz, nil

			case incentives.FutureRewardEstimate != nil:
				res, err := qp.GetFutureRewardsEstimate(ctx, incentives.FutureRewardEstimate.Owner, incentives.FutureRewardEstimate.StakeIDs, incentives.FutureRewardEstimate.NumEpochs)
				if err != nil {
					return nil, errors.Wrap(err, "unable to query FutureRewardEstimate")
				}

				bz, err := json.Marshal(res)
				if err != nil {
					return nil, errors.Wrap(err, "failed to JSON marshal FutureRewardEstimate response")
				}

				return bz, nil

			case incentives.AccountHistory != nil:
				res, err := qp.GetAccountHistory(ctx, incentives.AccountHistory.Account)
				if err != nil {
					return nil, errors.Wrap(err, "unable to query AccountHistory")
				}

				bz, err := json.Marshal(res)
				if err != nil {
					return nil, errors.Wrap(err, "failed to JSON marshal AccountHistory response")
				}

				return bz, nil

			case incentives.GaugeQualifyingValue != nil:
				res, err := qp.GetGaugeQualifyingValue(ctx, incentives.GaugeQualifyingValue.ID)
				if err != nil {
					return nil, errors.Wrap(err, "unable to query GaugeQualifyingValue")
				}

				bz, err := json.Marshal(res)
				if err != nil {
					return nil, errors.Wrap(err, "failed to JSON marshal GaugeQualifyingValue response")
				}

				return bz, nil

			default:
				return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown incentives neutron query type"}
			}

		default:
			return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown neutron query type"}
		}
	}
}
