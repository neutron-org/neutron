package wasmbinding

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"

	crontypes "github.com/neutron-org/neutron/v3/x/cron/types"

	dextypes "github.com/neutron-org/neutron/v3/x/dex/types"
	feeburnertypes "github.com/neutron-org/neutron/v3/x/feeburner/types"
	interchainqueriestypes "github.com/neutron-org/neutron/v3/x/interchainqueries/types"
	interchaintxstypes "github.com/neutron-org/neutron/v3/x/interchaintxs/types"
	tokenfactorytypes "github.com/neutron-org/neutron/v3/x/tokenfactory/types"
)

func AcceptedStargateQueries() wasmkeeper.AcceptedStargateQueries {
	return wasmkeeper.AcceptedStargateQueries{
		// ibc
		"/ibc.core.client.v1.Query/ClientState":    &ibcclienttypes.QueryClientStateResponse{},
		"/ibc.core.client.v1.Query/ConsensusState": &ibcclienttypes.QueryConsensusStateResponse{},
		"/ibc.core.connection.v1.Query/Connection": &ibcconnectiontypes.QueryConnectionResponse{},

		// token factory
		"/osmosis.tokenfactory.v1beta1.Query/Params":                 &tokenfactorytypes.QueryParamsResponse{},
		"/osmosis.tokenfactory.v1beta1.Query/DenomAuthorityMetadata": &tokenfactorytypes.QueryDenomAuthorityMetadataResponse{},
		"/osmosis.tokenfactory.v1beta1.Query/DenomsFromCreator":      &tokenfactorytypes.QueryDenomsFromCreatorResponse{},
		"/osmosis.tokenfactory.v1beta1.Query/BeforeSendHookAddress":  &tokenfactorytypes.QueryBeforeSendHookAddressResponse{},

		// interchain accounts
		"/ibc.applications.interchain_accounts.controller.v1.Query/InterchainAccount": &icacontrollertypes.QueryInterchainAccountResponse{},

		// transfer
		"/ibc.applications.transfer.v1.Query/DenomTrace": &ibctransfertypes.QueryDenomTraceResponse{},

		// auth
		"/cosmos.auth.v1beta1.Query/Account": &authtypes.QueryAccountResponse{},
		"/cosmos.auth.v1beta1.Query/Params":  &authtypes.QueryParamsResponse{},

		// bank
		"/cosmos.bank.v1beta1.Query/Balance":       &banktypes.QueryBalanceResponse{},
		"/cosmos.bank.v1beta1.Query/DenomMetadata": &banktypes.QueryDenomsMetadataResponse{},
		"/cosmos.bank.v1beta1.Query/Params":        &banktypes.QueryParamsResponse{},
		"/cosmos.bank.v1beta1.Query/SupplyOf":      &banktypes.QuerySupplyOfResponse{},

		// interchaintxs
		"/neutron.interchaintxs.v1.Query/Params":                   &interchaintxstypes.QueryParamsResponse{},
		"/neutron.interchaintxs.v1.Query/InterchainAccountAddress": &interchaintxstypes.QueryInterchainAccountAddressResponse{},

		// cron
		"/neutron.cron.Query/Params": &crontypes.QueryParamsResponse{},

		// interchainqueries
		"/neutron.interchainqueries.Query/Params":            &interchainqueriestypes.QueryParamsResponse{},
		"/neutron.interchainqueries.Query/RegisteredQueries": &interchainqueriestypes.QueryRegisteredQueriesResponse{},
		"/neutron.interchainqueries.Query/RegisteredQuery":   &interchainqueriestypes.QueryRegisteredQueryResponse{},
		"/neutron.interchainqueries.Query/QueryResult":       &interchainqueriestypes.QueryRegisteredQueryResultResponse{},
		"/neutron.interchainqueries.Query/LastRemoteHeight":  &interchainqueriestypes.QueryLastRemoteHeightResponse{},

		// feeburner
		"/neutron.feeburner.Query/Params":                    &feeburnertypes.QueryParamsResponse{},
		"/neutron.feeburner.Query/TotalBurnedNeutronsAmount": &feeburnertypes.QueryTotalBurnedNeutronsAmountResponse{},

		// dex
		"/neutron.dex.Query/Params":                            &dextypes.QueryParamsResponse{},
		"/neutron.dex.Query/LimitOrderTrancheUser":             &dextypes.QueryGetLimitOrderTrancheUserResponse{},
		"/neutron.dex.Query/LimitOrderTrancheUserAll":          &dextypes.QueryAllLimitOrderTrancheUserResponse{},
		"/neutron.dex.Query/LimitOrderTrancheUserAllByAddress": &dextypes.QueryAllUserLimitOrdersResponse{},
		"/neutron.dex.Query/LimitOrderTranche":                 &dextypes.QueryGetLimitOrderTrancheResponse{},
		"/neutron.dex.Query/LimitOrderTrancheAll":              &dextypes.QueryAllLimitOrderTrancheResponse{},
		"/neutron.dex.Query/UserDepositsAll":                   &dextypes.QueryAllUserDepositsResponse{},
		"/neutron.dex.Query/TickLiquidityAll":                  &dextypes.QueryAllTickLiquidityResponse{},
		"/neutron.dex.Query/UserLimitOrdersAll":                &dextypes.QueryAllUserLimitOrdersResponse{},
		"/neutron.dex.Query/InactiveLimitOrderTranche":         &dextypes.QueryGetInactiveLimitOrderTrancheResponse{},
		"/neutron.dex.Query/InactiveLimitOrderTrancheAll":      &dextypes.QueryAllInactiveLimitOrderTrancheResponse{},
		"/neutron.dex.Query/PoolReservesAll":                   &dextypes.QueryAllPoolReservesResponse{},
		"/neutron.dex.Query/PoolReserves":                      &dextypes.QueryGetPoolReservesResponse{},
		"/neutron.dex.Query/EstimateMultiHopSwap":              &dextypes.QueryEstimateMultiHopSwapResponse{},
		"/neutron.dex.Query/EstimatePlaceLimitOrder":           &dextypes.QueryEstimatePlaceLimitOrderResponse{},
		"/neutron.dex.Query/Pool":                              &dextypes.QueryPoolResponse{},
		"/neutron.dex.Query/PoolByID":                          &dextypes.QueryPoolResponse{},
		"/neutron.dex.Query/PoolMetadata":                      &dextypes.QueryGetPoolMetadataResponse{},
		"/neutron.dex.Query/PoolMetadataAll":                   &dextypes.QueryAllPoolMetadataResponse{},
	}
}
