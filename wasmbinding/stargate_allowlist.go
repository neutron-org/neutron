package wasmbinding

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
	epochstypes "github.com/neutron-org/neutron/x/epochs/types"
	feeburnertypes "github.com/neutron-org/neutron/x/feeburner/types"
	incentivestypes "github.com/neutron-org/neutron/x/incentives/types"
	interchainqueriestypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	interchaintxstypes "github.com/neutron-org/neutron/x/interchaintxs/types"
	tokenfactorytypes "github.com/neutron-org/neutron/x/tokenfactory/types"
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
		"/neutron.interchaintxs.Query/Params": &interchaintxstypes.QueryParamsResponse{},

		// interchainqueries
		"/neutron.interchainqueries.Query/Params": &interchainqueriestypes.QueryParamsResponse{},

		// feeburner
		"/neutron.feeburner.Query/Params": &feeburnertypes.QueryParamsResponse{},

		// dex
		"/dualitylabs.duality.dex.Query/Params":                    &dextypes.QueryParamsResponse{},
		"/dualitylabs.duality.dex.Query/LimitOrderTrancheUser":     &dextypes.QueryGetLimitOrderTrancheUserResponse{},
		"/dualitylabs.duality.dex.Query/LimitOrderTranche":         &dextypes.QueryGetLimitOrderTrancheResponse{},
		"/dualitylabs.duality.dex.Query/UserDepositsAll":           &dextypes.QueryAllUserDepositsResponse{},
		"/dualitylabs.duality.dex.Query/UserLimitOrdersAll":        &dextypes.QueryAllUserLimitOrdersResponse{},
		"/dualitylabs.duality.dex.Query/InactiveLimitOrderTranche": &dextypes.QueryGetInactiveLimitOrderTrancheResponse{},
		"/dualitylabs.duality.dex.Query/PoolReserves":              &dextypes.QueryGetPoolReservesResponse{},
		"/dualitylabs.duality.dex.Query/EstimateMultiHopSwap":      &dextypes.QueryEstimateMultiHopSwapResponse{},
		"/dualitylabs.duality.dex.Query/EstimatePlaceLimitOrder":   &dextypes.QueryEstimatePlaceLimitOrderResponse{},

		// incentives
		"/dualitylabs.duality.incentives.Query/GetModuleStatus":         &incentivestypes.GetModuleStatusResponse{},
		"/dualitylabs.duality.incentives.Query/GetGaugeByID":            &incentivestypes.GetGaugeByIDResponse{},
		"/dualitylabs.duality.incentives.Query/GetGauges":               &incentivestypes.GetGaugesResponse{},
		"/dualitylabs.duality.incentives.Query/GetStakeByID":            &incentivestypes.GetStakeByIDResponse{},
		"/dualitylabs.duality.incentives.Query/GetStakes":               &incentivestypes.GetStakesResponse{},
		"/dualitylabs.duality.incentives.Query/GetFutureRewardEstimate": &incentivestypes.GetFutureRewardEstimateResponse{},
		"/dualitylabs.duality.incentives.Query/GetAccountHistory":       &incentivestypes.GetAccountHistoryResponse{},
		"/dualitylabs.duality.incentives.Query/GetGaugeQualifyingValue": &incentivestypes.GetGaugeQualifyingValueResponse{},

		// epochs
		"/dualitylabs.duality.epochs.Query/EpochInfos":   &epochstypes.QueryEpochsInfoResponse{},
		"/dualitylabs.duality.epochs.Query/CurrentEpoch": &epochstypes.QueryCurrentEpochResponse{},
	}
}
