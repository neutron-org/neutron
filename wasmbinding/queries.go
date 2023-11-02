package wasmbinding

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	contractmanagertypes "github.com/neutron-org/neutron/x/contractmanager/types"
	"github.com/neutron-org/neutron/x/incentives/keeper"
	incentivestypes "github.com/neutron-org/neutron/x/incentives/types"

	"github.com/neutron-org/neutron/wasmbinding/bindings"
	"github.com/neutron-org/neutron/x/interchainqueries/types"
	icatypes "github.com/neutron-org/neutron/x/interchaintxs/types"
)

func (qp *QueryPlugin) GetInterchainQueryResult(ctx sdk.Context, queryID uint64) (*bindings.QueryRegisteredQueryResultResponse, error) {
	grpcResp, err := qp.icqKeeper.GetQueryResultByID(ctx, queryID)
	if err != nil {
		return nil, err
	}
	resp := bindings.QueryResult{
		KvResults: make([]*bindings.StorageValue, 0, len(grpcResp.KvResults)),
		Height:    grpcResp.GetHeight(),
		Revision:  grpcResp.GetRevision(),
	}
	for _, grpcKv := range grpcResp.GetKvResults() {
		kv := bindings.StorageValue{
			StoragePrefix: grpcKv.GetStoragePrefix(),
			Key:           grpcKv.GetKey(),
			Value:         grpcKv.GetValue(),
		}
		resp.KvResults = append(resp.KvResults, &kv)
	}

	return &bindings.QueryRegisteredQueryResultResponse{Result: &resp}, nil
}

func (qp *QueryPlugin) GetInterchainAccountAddress(ctx sdk.Context, req *bindings.QueryInterchainAccountAddressRequest) (*bindings.QueryInterchainAccountAddressResponse, error) {
	grpcReq := icatypes.QueryInterchainAccountAddressRequest{
		OwnerAddress:        req.OwnerAddress,
		InterchainAccountId: req.InterchainAccountID,
		ConnectionId:        req.ConnectionID,
	}

	grpcResp, err := qp.icaControllerKeeper.InterchainAccountAddress(sdk.WrapSDKContext(ctx), &grpcReq)
	if err != nil {
		return nil, err
	}

	return &bindings.QueryInterchainAccountAddressResponse{InterchainAccountAddress: grpcResp.GetInterchainAccountAddress()}, nil
}

func (qp *QueryPlugin) GetRegisteredInterchainQueries(ctx sdk.Context, query *bindings.QueryRegisteredQueriesRequest) (*bindings.QueryRegisteredQueriesResponse, error) {
	grpcResp, err := qp.icqKeeper.GetRegisteredQueries(ctx, &types.QueryRegisteredQueriesRequest{
		Owners:       query.Owners,
		ConnectionId: query.ConnectionID,
		Pagination: &sdkquery.PageRequest{
			Key:        query.Pagination.Key,
			Offset:     query.Pagination.Offset,
			Limit:      query.Pagination.Limit,
			CountTotal: query.Pagination.CountTotal,
			Reverse:    query.Pagination.Reverse,
		},
	})
	if err != nil {
		return nil, err
	}

	resp := bindings.QueryRegisteredQueriesResponse{RegisteredQueries: make([]bindings.RegisteredQuery, 0, len(grpcResp.GetRegisteredQueries()))}
	for _, grpcQuery := range grpcResp.GetRegisteredQueries() {
		query := mapGRPCRegisteredQueryToWasmBindings(grpcQuery)
		resp.RegisteredQueries = append(resp.RegisteredQueries, query)
	}
	return &resp, nil
}

func (qp *QueryPlugin) GetRegisteredInterchainQuery(ctx sdk.Context, req *bindings.QueryRegisteredQueryRequest) (*bindings.QueryRegisteredQueryResponse, error) {
	grpcResp, err := qp.icqKeeper.GetQueryByID(ctx, req.QueryID)
	if err != nil {
		return nil, err
	}
	if grpcResp == nil {
		return nil, errors.Wrapf(types.ErrEmptyResult, "interchain query response empty for query id %d", req.QueryID)
	}
	query := mapGRPCRegisteredQueryToWasmBindings(*grpcResp)

	return &bindings.QueryRegisteredQueryResponse{RegisteredQuery: &query}, nil
}

// GetDenomAdmin is a query to get denom admin.
func (qp QueryPlugin) GetDenomAdmin(ctx sdk.Context, denom string) (*bindings.DenomAdminResponse, error) {
	metadata, err := qp.tokenFactoryKeeper.GetAuthorityMetadata(ctx, denom)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get admin for denom: %s", denom)
	}

	return &bindings.DenomAdminResponse{Admin: metadata.Admin}, nil
}

// GetBeforeSendHook is a query to get denom before send hook.
func (qp QueryPlugin) GetBeforeSendHook(ctx sdk.Context, denom string) (*bindings.BeforeSendHookResponse, error) {
	contractAddr := qp.tokenFactoryKeeper.GetBeforeSendHook(ctx, denom)

	return &bindings.BeforeSendHookResponse{ContractAddr: contractAddr}, nil
}

func (qp *QueryPlugin) GetTotalBurnedNeutronsAmount(ctx sdk.Context, _ *bindings.QueryTotalBurnedNeutronsAmountRequest) (*bindings.QueryTotalBurnedNeutronsAmountResponse, error) {
	grpcResp := qp.feeBurnerKeeper.GetTotalBurnedNeutronsAmount(ctx)
	return &bindings.QueryTotalBurnedNeutronsAmountResponse{Coin: grpcResp.Coin}, nil
}

func (qp *QueryPlugin) GetMinIbcFee(ctx sdk.Context, _ *bindings.QueryMinIbcFeeRequest) (*bindings.QueryMinIbcFeeResponse, error) {
	fee := qp.feeRefunderKeeper.GetMinFee(ctx)
	return &bindings.QueryMinIbcFeeResponse{MinFee: fee}, nil
}

func (qp *QueryPlugin) GetFailures(ctx sdk.Context, address string, pagination *sdkquery.PageRequest) (*bindings.FailuresResponse, error) {
	res, err := qp.contractmanagerKeeper.AddressFailures(ctx, &contractmanagertypes.QueryFailuresRequest{
		Address:    address,
		Pagination: pagination,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get failures for address: %s", address)
	}

	return &bindings.FailuresResponse{Failures: res.Failures}, nil
}

func (qp *QueryPlugin) GetGauges(ctx sdk.Context, status string, denom string) (*bindings.GaugesResponse, error) {
	queryServer := keeper.NewQueryServer(qp.incentivesKeeper)
	gaugeStatusInt, ok := incentivestypes.GaugeStatus_value[status]
	if !ok {
		return nil, errors.Wrapf(incentivestypes.ErrInvalidGaugeStatus, "failed to parse gauge status")
	}
	res, err := queryServer.GetGauges(ctx, &incentivestypes.GetGaugesRequest{
		Status: incentivestypes.GaugeStatus(gaugeStatusInt),
		Denom:  denom,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get gauges for status=%s denom=%s", status, denom)
	}

	return &bindings.GaugesResponse{
		Gauges: res.Gauges,
	}, nil
}

func (qp *QueryPlugin) GetModuleStatus(ctx sdk.Context) (*bindings.ModuleStatusResponse, error) {
	queryServer := keeper.NewQueryServer(qp.incentivesKeeper)
	res, err := queryServer.GetModuleStatus(ctx, &incentivestypes.GetModuleStatusRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get module status")
	}

	return &bindings.ModuleStatusResponse{
		RewardCoins: res.RewardCoins,
		Params:      res.Params,
	}, nil
}

func (qp *QueryPlugin) GetGaugeByID(ctx sdk.Context, id uint64) (*bindings.GaugeByIDResponse, error) {
	queryServer := keeper.NewQueryServer(qp.incentivesKeeper)
	res, err := queryServer.GetGaugeByID(ctx, &incentivestypes.GetGaugeByIDRequest{
		Id: id,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get gauge for ID=%d", id)
	}

	return &bindings.GaugeByIDResponse{
		Gauge: res.Gauge,
	}, nil
}

func (qp *QueryPlugin) GetStakeByID(ctx sdk.Context, id uint64) (*bindings.StakeByIDResponse, error) {
	queryServer := keeper.NewQueryServer(qp.incentivesKeeper)
	res, err := queryServer.GetStakeByID(ctx, &incentivestypes.GetStakeByIDRequest{
		StakeId: id,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get stake for ID=%d", id)
	}

	return &bindings.StakeByIDResponse{
		Stake: res.Stake,
	}, nil
}

func (qp *QueryPlugin) GetStakes(ctx sdk.Context, owner string) (*bindings.StakesResponse, error) {
	queryServer := keeper.NewQueryServer(qp.incentivesKeeper)
	res, err := queryServer.GetStakes(ctx, &incentivestypes.GetStakesRequest{
		Owner: owner,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get stakes for owner=%s", owner)
	}

	return &bindings.StakesResponse{
		Stakes: res.Stakes,
	}, nil
}

func (qp *QueryPlugin) GetFutureRewardsEstimate(ctx sdk.Context, owner string, stakeIds []uint64, numEpochs int64) (*bindings.FutureRewardEstimateResponse, error) {
	queryServer := keeper.NewQueryServer(qp.incentivesKeeper)
	res, err := queryServer.GetFutureRewardEstimate(ctx, &incentivestypes.GetFutureRewardEstimateRequest{
		Owner:     owner,
		StakeIds:  stakeIds,
		NumEpochs: numEpochs,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get future rewards estimate for owner=%s", owner)
	}

	return &bindings.FutureRewardEstimateResponse{
		Coins: res.Coins,
	}, nil
}

func (qp *QueryPlugin) GetAccountHistory(ctx sdk.Context, account string) (*bindings.AccountHistoryResponse, error) {
	queryServer := keeper.NewQueryServer(qp.incentivesKeeper)
	res, err := queryServer.GetAccountHistory(ctx, &incentivestypes.GetAccountHistoryRequest{
		Account: account,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get account history for account=%s", account)
	}

	return &bindings.AccountHistoryResponse{
		Coins: res.Coins,
	}, nil
}

func (qp *QueryPlugin) GetGaugeQualifyingValue(ctx sdk.Context, id uint64) (*bindings.GaugeQualifyingValueResponse, error) {
	queryServer := keeper.NewQueryServer(qp.incentivesKeeper)
	res, err := queryServer.GetGaugeQualifyingValue(ctx, &incentivestypes.GetGaugeQualifyingValueRequest{
		Id: id,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get gauge qualifying value for id=%d", id)
	}

	return &bindings.GaugeQualifyingValueResponse{
		QualifyingValue: res.QualifyingValue,
	}, nil
}

func mapGRPCRegisteredQueryToWasmBindings(grpcQuery types.RegisteredQuery) bindings.RegisteredQuery {
	return bindings.RegisteredQuery{
		ID:                              grpcQuery.GetId(),
		Owner:                           grpcQuery.GetOwner(),
		Keys:                            grpcQuery.GetKeys(),
		TransactionsFilter:              grpcQuery.GetTransactionsFilter(),
		QueryType:                       grpcQuery.GetQueryType(),
		ConnectionID:                    grpcQuery.GetConnectionId(),
		UpdatePeriod:                    grpcQuery.GetUpdatePeriod(),
		LastSubmittedResultLocalHeight:  grpcQuery.GetLastSubmittedResultLocalHeight(),
		LastSubmittedResultRemoteHeight: grpcQuery.GetLastSubmittedResultRemoteHeight(),
		Deposit:                         grpcQuery.GetDeposit(),
		SubmitTimeout:                   grpcQuery.GetSubmitTimeout(),
		RegisteredAtHeight:              grpcQuery.GetRegisteredAtHeight(),
	}
}
