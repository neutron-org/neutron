package wasmbinding

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/exp/maps"

	"cosmossdk.io/errors"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"

	dextypes "github.com/neutron-org/neutron/v5/x/dex/types"

	contractmanagertypes "github.com/neutron-org/neutron/v5/x/contractmanager/types"

	marketmapkeeper "github.com/skip-mev/connect/v2/x/marketmap/keeper"
	oraclekeeper "github.com/skip-mev/connect/v2/x/oracle/keeper"

	"github.com/neutron-org/neutron/v5/wasmbinding/bindings"
	"github.com/neutron-org/neutron/v5/x/interchainqueries/types"
	icatypes "github.com/neutron-org/neutron/v5/x/interchaintxs/types"
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

	grpcResp, err := qp.icaControllerKeeper.InterchainAccountAddress(ctx, &grpcReq)
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

func (qp *QueryPlugin) DexQuery(ctx sdk.Context, query bindings.DexQuery) (data []byte, err error) {
	switch {
	case query.EstimateMultiHopSwap != nil:
		data, err = dexQuery(ctx, query.EstimateMultiHopSwap, qp.dexKeeper.EstimateMultiHopSwap)
	case query.EstimatePlaceLimitOrder != nil:
		q := dextypes.QueryEstimatePlaceLimitOrderRequest{
			Creator:          query.EstimatePlaceLimitOrder.Creator,
			Receiver:         query.EstimatePlaceLimitOrder.Receiver,
			TokenIn:          query.EstimatePlaceLimitOrder.TokenIn,
			TokenOut:         query.EstimatePlaceLimitOrder.TokenOut,
			TickIndexInToOut: query.EstimatePlaceLimitOrder.TickIndexInToOut,
			AmountIn:         query.EstimatePlaceLimitOrder.AmountIn,
			MaxAmountOut:     query.EstimatePlaceLimitOrder.MaxAmountOut,
		}
		orderTypeInt, ok := dextypes.LimitOrderType_value[query.EstimatePlaceLimitOrder.OrderType]
		if !ok {
			return nil, errors.Wrap(dextypes.ErrInvalidOrderType,
				fmt.Sprintf(
					"got \"%s\", expected one of %s",
					query.EstimatePlaceLimitOrder.OrderType,
					strings.Join(maps.Keys(dextypes.LimitOrderType_value), ", ")),
			)
		}
		q.OrderType = dextypes.LimitOrderType(orderTypeInt)
		if query.EstimatePlaceLimitOrder.ExpirationTime != nil {
			t := time.Unix(int64(*query.EstimatePlaceLimitOrder.ExpirationTime), 0)
			q.ExpirationTime = &t
		}
		data, err = dexQuery(ctx, &q, qp.dexKeeper.EstimatePlaceLimitOrder)
	case query.InactiveLimitOrderTranche != nil:
		data, err = dexQuery(ctx, query.InactiveLimitOrderTranche, qp.dexKeeper.InactiveLimitOrderTranche)
	case query.InactiveLimitOrderTrancheAll != nil:
		data, err = dexQuery(ctx, query.InactiveLimitOrderTrancheAll, qp.dexKeeper.InactiveLimitOrderTrancheAll)
	case query.LimitOrderTrancheUser != nil:
		data, err = dexQuery(ctx, query.LimitOrderTrancheUser, qp.dexKeeper.LimitOrderTrancheUser)
	case query.LimitOrderTranche != nil:
		data, err = dexQuery(ctx, query.LimitOrderTranche, qp.dexKeeper.LimitOrderTranche)
	case query.LimitOrderTrancheAll != nil:
		data, err = dexQuery(ctx, query.LimitOrderTrancheAll, qp.dexKeeper.LimitOrderTrancheAll)
	case query.LimitOrderTrancheUserAll != nil:
		data, err = dexQuery(ctx, query.LimitOrderTrancheUserAll, qp.dexKeeper.LimitOrderTrancheUserAll)
	case query.Params != nil:
		data, err = dexQuery(ctx, query.Params, qp.dexKeeper.Params)
	case query.Pool != nil:
		data, err = dexQuery(ctx, query.Pool, qp.dexKeeper.Pool)
	case query.PoolByID != nil:
		data, err = dexQuery(ctx, query.PoolByID, qp.dexKeeper.PoolByID)
	case query.LimitOrderTrancheUserAllByAddress != nil:
		data, err = dexQuery(ctx, query.LimitOrderTrancheUserAllByAddress, qp.dexKeeper.LimitOrderTrancheUserAllByAddress)
	case query.PoolMetadata != nil:
		data, err = dexQuery(ctx, query.PoolMetadata, qp.dexKeeper.PoolMetadata)
	case query.PoolMetadataAll != nil:
		data, err = dexQuery(ctx, query.PoolMetadataAll, qp.dexKeeper.PoolMetadataAll)
	case query.PoolReservesAll != nil:
		data, err = dexQuery(ctx, query.PoolReservesAll, qp.dexKeeper.PoolReservesAll)
	case query.PoolReserves != nil:
		data, err = dexQuery(ctx, query.PoolReserves, qp.dexKeeper.PoolReserves)
	case query.TickLiquidityAll != nil:
		data, err = dexQuery(ctx, query.TickLiquidityAll, qp.dexKeeper.TickLiquidityAll)
	case query.UserDepositsAll != nil:
		data, err = dexQuery(ctx, query.UserDepositsAll, qp.dexKeeper.UserDepositsAll)
	default:
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown neutron.dex query type"}
	}

	return data, err
}

func (qp *QueryPlugin) OracleQuery(ctx sdk.Context, query bindings.OracleQuery) ([]byte, error) {
	oracleQueryServer := oraclekeeper.NewQueryServer(*qp.oracleKeeper)

	switch {
	case query.GetAllCurrencyPairs != nil:
		return processResponse(oracleQueryServer.GetAllCurrencyPairs(ctx, query.GetAllCurrencyPairs))
	case query.GetPrice != nil:
		return processResponse(oracleQueryServer.GetPrice(ctx, query.GetPrice))
	case query.GetPrices != nil:
		return processResponse(oracleQueryServer.GetPrices(ctx, query.GetPrices))
	default:
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown neutron.oracle query type"}
	}
}

func (qp *QueryPlugin) MarketMapQuery(ctx sdk.Context, query bindings.MarketMapQuery) ([]byte, error) {
	marketMapQueryServer := marketmapkeeper.NewQueryServer(qp.marketmapKeeper)

	switch {
	case query.Params != nil:
		return processResponse(marketMapQueryServer.Params(ctx, query.Params))
	case query.LastUpdated != nil:
		return processResponse(marketMapQueryServer.LastUpdated(ctx, query.LastUpdated))
	case query.MarketMap != nil:
		return processResponse(marketMapQueryServer.MarketMap(ctx, query.MarketMap))
	case query.Market != nil:
		return processResponse(marketMapQueryServer.Market(ctx, query.Market))
	default:
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown neutron.marketmap query type"}
	}
}

func dexQuery[T, R any](ctx sdk.Context, query *T, queryHandler func(ctx context.Context, query *T) (R, error)) ([]byte, error) {
	resp, err := queryHandler(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("failed to query request %T", query))
	}
	var data []byte

	if q, ok := any(resp).(bindings.BindingMarshaller); ok {
		data, err = q.MarshalBinding()
	} else {
		data, err = json.Marshal(resp)
	}

	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("failed to marshal response %T", resp))
	}

	return data, nil
}

func processResponse(resp interface{}, err error) ([]byte, error) {
	if err != nil {
		return nil, errors.Wrapf(err, "failed to process request %T", resp)
	}
	if q, ok := resp.(bindings.BindingMarshaller); ok {
		return q.MarshalBinding()
	}
	return json.Marshal(resp)
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
