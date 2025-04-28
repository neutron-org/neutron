package wasmbinding

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types" //nolint:staticcheck
	ibcconnectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	consumertypes "github.com/cosmos/interchain-security/v5/x/ccv/consumer/types"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
	marketmaptypes "github.com/skip-mev/slinky/x/marketmap/types"
	oracletypes "github.com/skip-mev/slinky/x/oracle/types"

	harpoontypes "github.com/neutron-org/neutron/v6/x/harpoon/types"

	globalfeetypes "github.com/neutron-org/neutron/v6/x/globalfee/types"

	dynamicfeestypes "github.com/neutron-org/neutron/v6/x/dynamicfees/types"

	crontypes "github.com/neutron-org/neutron/v6/x/cron/types"
	dextypes "github.com/neutron-org/neutron/v6/x/dex/types"
	feeburnertypes "github.com/neutron-org/neutron/v6/x/feeburner/types"
	interchainqueriestypes "github.com/neutron-org/neutron/v6/x/interchainqueries/types"
	interchaintxstypes "github.com/neutron-org/neutron/v6/x/interchaintxs/types"
	tokenfactorytypes "github.com/neutron-org/neutron/v6/x/tokenfactory/types"
)

func AcceptedStargateQueries() wasmkeeper.AcceptedQueries {
	return wasmkeeper.AcceptedQueries{
		// ibc
		"/ibc.core.client.v1.Query/ClientState":         &ibcclienttypes.QueryClientStateResponse{},
		"/ibc.core.client.v1.Query/ConsensusState":      &ibcclienttypes.QueryConsensusStateResponse{},
		"/ibc.core.connection.v1.Query/Connection":      &ibcconnectiontypes.QueryConnectionResponse{},
		"/ibc.core.channel.v1.Query/ChannelClientState": &ibcchanneltypes.QueryChannelClientStateResponse{},

		// token factory
		"/osmosis.tokenfactory.v1beta1.Query/Params":                 &tokenfactorytypes.QueryParamsResponse{},
		"/osmosis.tokenfactory.v1beta1.Query/DenomAuthorityMetadata": &tokenfactorytypes.QueryDenomAuthorityMetadataResponse{},
		"/osmosis.tokenfactory.v1beta1.Query/DenomsFromCreator":      &tokenfactorytypes.QueryDenomsFromCreatorResponse{},
		"/osmosis.tokenfactory.v1beta1.Query/BeforeSendHookAddress":  &tokenfactorytypes.QueryBeforeSendHookAddressResponse{},
		"/osmosis.tokenfactory.v1beta1.Query/FullDenom":              &tokenfactorytypes.QueryFullDenomResponse{},

		// interchain accounts
		"/ibc.applications.interchain_accounts.controller.v1.Query/InterchainAccount": &icacontrollertypes.QueryInterchainAccountResponse{},

		// transfer
		"/ibc.applications.transfer.v1.Query/DenomTrace":    &ibctransfertypes.QueryDenomTraceResponse{},
		"/ibc.applications.transfer.v1.Query/EscrowAddress": &ibctransfertypes.QueryEscrowAddressResponse{},

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
		"/neutron.dex.Query/LimitOrderTrancheUserAllByAddress": &dextypes.QueryAllLimitOrderTrancheUserByAddressResponse{},
		"/neutron.dex.Query/LimitOrderTranche":                 &dextypes.QueryGetLimitOrderTrancheResponse{},
		"/neutron.dex.Query/LimitOrderTrancheAll":              &dextypes.QueryAllLimitOrderTrancheResponse{},
		"/neutron.dex.Query/UserDepositsAll":                   &dextypes.QueryAllUserDepositsResponse{},
		"/neutron.dex.Query/TickLiquidityAll":                  &dextypes.QueryAllTickLiquidityResponse{},
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
		"/neutron.dex.Query/SimulateDeposit":                   &dextypes.QuerySimulateDepositResponse{},
		"/neutron.dex.Query/SimulateWithdrawal":                &dextypes.QuerySimulateWithdrawalResponse{},
		"/neutron.dex.Query/SimulatePlaceLimitOrder":           &dextypes.QuerySimulatePlaceLimitOrderResponse{},
		"/neutron.dex.Query/SimulateWithdrawFilledLimitOrder":  &dextypes.QuerySimulateWithdrawFilledLimitOrderResponse{},
		"/neutron.dex.Query/SimulateCancelLimitOrder":          &dextypes.QuerySimulateCancelLimitOrderResponse{},
		"/neutron.dex.Query/SimulateMultiHopSwap":              &dextypes.QuerySimulateMultiHopSwapResponse{},

		// oracle
		"/slinky.oracle.v1.Query/GetAllCurrencyPairs": &oracletypes.GetAllCurrencyPairsResponse{},
		"/slinky.oracle.v1.Query/GetPrice":            &oracletypes.GetPriceResponse{},
		"/slinky.oracle.v1.Query/GetPrices":           &oracletypes.GetPricesResponse{},

		// marketmap
		"/slinky.marketmap.v1.Query/Markets":     &marketmaptypes.MarketsResponse{},
		"/slinky.marketmap.v1.Query/LastUpdated": &marketmaptypes.LastUpdatedResponse{},
		"/slinky.marketmap.v1.Query/Params":      &marketmaptypes.ParamsResponse{},
		"/slinky.marketmap.v1.Query/Market":      &marketmaptypes.MarketResponse{},

		// feemarket
		"/feemarket.feemarket.v1.Query/Params":    &feemarkettypes.ParamsResponse{},
		"/feemarket.feemarket.v1.Query/State":     &feemarkettypes.StateResponse{},
		"/feemarket.feemarket.v1.Query/GasPrice":  &feemarkettypes.GasPriceResponse{},
		"/feemarket.feemarket.v1.Query/GasPrices": &feemarkettypes.GasPricesResponse{},

		// dynamicfees
		"/neutron.dynamicfees.v1.Query/Params": &dynamicfeestypes.QueryParamsResponse{},

		// globalfee
		"/gaia.globalfee.v1beta1.Query/Params": &globalfeetypes.QueryParamsResponse{},

		// consumer
		"/interchain_security.ccv.consumer.v1.Query/QueryParams": &consumertypes.QueryParamsResponse{},

		// distribution
		"/cosmos.distribution.v1beta1.Query/DelegationRewards": &types.QueryDelegationRewardsResponse{},

		// staking
		"/cosmos.staking.v1beta1.Query/Delegation":                    &stakingtypes.QueryDelegationResponse{},
		"/cosmos.staking.v1beta1.Query/UnbondingDelegation":           &stakingtypes.QueryUnbondingDelegationResponse{},
		"/cosmos.staking.v1beta1.Query/Validator":                     &stakingtypes.QueryValidatorResponse{},
		"/cosmos.staking.v1beta1.Query/DelegatorDelegations":          &stakingtypes.QueryDelegatorDelegationsResponse{},
		"/cosmos.staking.v1beta1.Query/DelegatorUnbondingDelegations": &stakingtypes.QueryDelegatorUnbondingDelegationsResponse{},

		// harpoon
		"/neutron.harpoon.Query/SubscribedContracts": &harpoontypes.QuerySubscribedContractsResponse{},
	}
}
