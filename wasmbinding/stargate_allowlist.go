package wasmbinding

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
	marketmaptypes "github.com/skip-mev/slinky/x/marketmap/types"
	oracletypes "github.com/skip-mev/slinky/x/oracle/types"

	crontypes "github.com/neutron-org/neutron/v8/x/cron/types"
	dextypes "github.com/neutron-org/neutron/v8/x/dex/types"
	feeburnertypes "github.com/neutron-org/neutron/v8/x/feeburner/types"
	interchainqueriestypes "github.com/neutron-org/neutron/v8/x/interchainqueries/types"
	interchaintxstypes "github.com/neutron-org/neutron/v8/x/interchaintxs/types"
	stateverifiertypes "github.com/neutron-org/neutron/v8/x/state-verifier/types"
	tokenfactorytypes "github.com/neutron-org/neutron/v8/x/tokenfactory/types"

	harpoontypes "github.com/neutron-org/neutron/v8/x/harpoon/types"

	globalfeetypes "github.com/neutron-org/neutron/v8/x/globalfee/types"

	dynamicfeestypes "github.com/neutron-org/neutron/v8/x/dynamicfees/types"
)

func AcceptedStargateQueries() wasmkeeper.AcceptedQueries {
	return wasmkeeper.AcceptedQueries{
		// ibc
		"/ibc.core.client.v1.Query/ClientState":         func() proto.Message { return &ibcclienttypes.QueryClientStateResponse{} },
		"/ibc.core.client.v1.Query/ConsensusState":      func() proto.Message { return &ibcclienttypes.QueryConsensusStateResponse{} },
		"/ibc.core.connection.v1.Query/Connection":      func() proto.Message { return &ibcconnectiontypes.QueryConnectionResponse{} },
		"/ibc.core.channel.v1.Query/ChannelClientState": func() proto.Message { return &ibcchanneltypes.QueryChannelClientStateResponse{} },

		// token factory
		"/osmosis.tokenfactory.v1beta1.Query/Params":                 func() proto.Message { return &tokenfactorytypes.QueryParamsResponse{} },
		"/osmosis.tokenfactory.v1beta1.Query/DenomAuthorityMetadata": func() proto.Message { return &tokenfactorytypes.QueryDenomAuthorityMetadataResponse{} },
		"/osmosis.tokenfactory.v1beta1.Query/DenomsFromCreator":      func() proto.Message { return &tokenfactorytypes.QueryDenomsFromCreatorResponse{} },
		"/osmosis.tokenfactory.v1beta1.Query/BeforeSendHookAddress":  func() proto.Message { return &tokenfactorytypes.QueryBeforeSendHookAddressResponse{} },
		"/osmosis.tokenfactory.v1beta1.Query/FullDenom":              func() proto.Message { return &tokenfactorytypes.QueryFullDenomResponse{} },

		// interchain accounts
		"/ibc.applications.interchain_accounts.controller.v1.Query/InterchainAccount": func() proto.Message { return &icacontrollertypes.QueryInterchainAccountResponse{} },

		// transfer
		"/ibc.applications.transfer.v1.Query/DenomTrace":    func() proto.Message { return &ibctransfertypes.QueryDenomTraceResponse{} },
		"/ibc.applications.transfer.v1.Query/EscrowAddress": func() proto.Message { return &ibctransfertypes.QueryEscrowAddressResponse{} },

		// auth
		"/cosmos.auth.v1beta1.Query/Account": func() proto.Message { return &authtypes.QueryAccountResponse{} },
		"/cosmos.auth.v1beta1.Query/Params":  func() proto.Message { return &authtypes.QueryParamsResponse{} },

		// bank
		"/cosmos.bank.v1beta1.Query/Balance":       func() proto.Message { return &banktypes.QueryBalanceResponse{} },
		"/cosmos.bank.v1beta1.Query/DenomMetadata": func() proto.Message { return &banktypes.QueryDenomsMetadataResponse{} },
		"/cosmos.bank.v1beta1.Query/Params":        func() proto.Message { return &banktypes.QueryParamsResponse{} },
		"/cosmos.bank.v1beta1.Query/SupplyOf":      func() proto.Message { return &banktypes.QuerySupplyOfResponse{} },

		// interchaintxs
		"/neutron.interchaintxs.v1.Query/Params":                   func() proto.Message { return &interchaintxstypes.QueryParamsResponse{} },
		"/neutron.interchaintxs.v1.Query/InterchainAccountAddress": func() proto.Message { return &interchaintxstypes.QueryInterchainAccountAddressResponse{} },

		// cron
		"/neutron.cron.Query/Params": func() proto.Message { return &crontypes.QueryParamsResponse{} },

		// interchainqueries
		"/neutron.interchainqueries.Query/Params":            func() proto.Message { return &interchainqueriestypes.QueryParamsResponse{} },
		"/neutron.interchainqueries.Query/RegisteredQueries": func() proto.Message { return &interchainqueriestypes.QueryRegisteredQueriesResponse{} },
		"/neutron.interchainqueries.Query/RegisteredQuery":   func() proto.Message { return &interchainqueriestypes.QueryRegisteredQueryResponse{} },
		"/neutron.interchainqueries.Query/QueryResult":       func() proto.Message { return &interchainqueriestypes.QueryRegisteredQueryResultResponse{} },
		"/neutron.interchainqueries.Query/LastRemoteHeight":  func() proto.Message { return &interchainqueriestypes.QueryLastRemoteHeightResponse{} },

		// feeburner
		"/neutron.feeburner.Query/Params":                    func() proto.Message { return &feeburnertypes.QueryParamsResponse{} },
		"/neutron.feeburner.Query/TotalBurnedNeutronsAmount": func() proto.Message { return &feeburnertypes.QueryTotalBurnedNeutronsAmountResponse{} },

		// dex
		"/neutron.dex.Query/Params":                            func() proto.Message { return &dextypes.QueryParamsResponse{} },
		"/neutron.dex.Query/LimitOrderTrancheUser":             func() proto.Message { return &dextypes.QueryGetLimitOrderTrancheUserResponse{} },
		"/neutron.dex.Query/LimitOrderTrancheUserAll":          func() proto.Message { return &dextypes.QueryAllLimitOrderTrancheUserResponse{} },
		"/neutron.dex.Query/LimitOrderTrancheUserAllByAddress": func() proto.Message { return &dextypes.QueryAllLimitOrderTrancheUserByAddressResponse{} },
		"/neutron.dex.Query/LimitOrderTranche":                 func() proto.Message { return &dextypes.QueryGetLimitOrderTrancheResponse{} },
		"/neutron.dex.Query/LimitOrderTrancheAll":              func() proto.Message { return &dextypes.QueryAllLimitOrderTrancheResponse{} },
		"/neutron.dex.Query/UserDepositsAll":                   func() proto.Message { return &dextypes.QueryAllUserDepositsResponse{} },
		"/neutron.dex.Query/TickLiquidityAll":                  func() proto.Message { return &dextypes.QueryAllTickLiquidityResponse{} },
		"/neutron.dex.Query/InactiveLimitOrderTranche":         func() proto.Message { return &dextypes.QueryGetInactiveLimitOrderTrancheResponse{} },
		"/neutron.dex.Query/InactiveLimitOrderTrancheAll":      func() proto.Message { return &dextypes.QueryAllInactiveLimitOrderTrancheResponse{} },
		"/neutron.dex.Query/PoolReservesAll":                   func() proto.Message { return &dextypes.QueryAllPoolReservesResponse{} },
		"/neutron.dex.Query/PoolReserves":                      func() proto.Message { return &dextypes.QueryGetPoolReservesResponse{} },
		"/neutron.dex.Query/EstimateMultiHopSwap":              func() proto.Message { return &dextypes.QueryEstimateMultiHopSwapResponse{} },
		"/neutron.dex.Query/EstimatePlaceLimitOrder":           func() proto.Message { return &dextypes.QueryEstimatePlaceLimitOrderResponse{} },
		"/neutron.dex.Query/Pool":                              func() proto.Message { return &dextypes.QueryPoolResponse{} },
		"/neutron.dex.Query/PoolByID":                          func() proto.Message { return &dextypes.QueryPoolResponse{} },
		"/neutron.dex.Query/PoolMetadata":                      func() proto.Message { return &dextypes.QueryGetPoolMetadataResponse{} },
		"/neutron.dex.Query/PoolMetadataAll":                   func() proto.Message { return &dextypes.QueryAllPoolMetadataResponse{} },
		"/neutron.dex.Query/SimulateDeposit":                   func() proto.Message { return &dextypes.QuerySimulateDepositResponse{} },
		"/neutron.dex.Query/SimulateWithdrawal":                func() proto.Message { return &dextypes.QuerySimulateWithdrawalResponse{} },
		"/neutron.dex.Query/SimulatePlaceLimitOrder":           func() proto.Message { return &dextypes.QuerySimulatePlaceLimitOrderResponse{} },
		"/neutron.dex.Query/SimulateWithdrawFilledLimitOrder":  func() proto.Message { return &dextypes.QuerySimulateWithdrawFilledLimitOrderResponse{} },
		"/neutron.dex.Query/SimulateCancelLimitOrder":          func() proto.Message { return &dextypes.QuerySimulateCancelLimitOrderResponse{} },
		"/neutron.dex.Query/SimulateMultiHopSwap":              func() proto.Message { return &dextypes.QuerySimulateMultiHopSwapResponse{} },

		// oracle
		"/slinky.oracle.v1.Query/GetAllCurrencyPairs": func() proto.Message { return &oracletypes.GetAllCurrencyPairsResponse{} },
		"/slinky.oracle.v1.Query/GetPrice":            func() proto.Message { return &oracletypes.GetPriceResponse{} },
		"/slinky.oracle.v1.Query/GetPrices":           func() proto.Message { return &oracletypes.GetPricesResponse{} },

		// marketmap
		"/slinky.marketmap.v1.Query/Markets":     func() proto.Message { return &marketmaptypes.MarketsResponse{} },
		"/slinky.marketmap.v1.Query/LastUpdated": func() proto.Message { return &marketmaptypes.LastUpdatedResponse{} },
		"/slinky.marketmap.v1.Query/Params":      func() proto.Message { return &marketmaptypes.ParamsResponse{} },
		"/slinky.marketmap.v1.Query/Market":      func() proto.Message { return &marketmaptypes.MarketResponse{} },

		// feemarket
		"/feemarket.feemarket.v1.Query/Params":    func() proto.Message { return &feemarkettypes.ParamsResponse{} },
		"/feemarket.feemarket.v1.Query/State":     func() proto.Message { return &feemarkettypes.StateResponse{} },
		"/feemarket.feemarket.v1.Query/GasPrice":  func() proto.Message { return &feemarkettypes.GasPriceResponse{} },
		"/feemarket.feemarket.v1.Query/GasPrices": func() proto.Message { return &feemarkettypes.GasPricesResponse{} },

		// dynamicfees
		"/neutron.dynamicfees.v1.Query/Params": func() proto.Message { return &dynamicfeestypes.QueryParamsResponse{} },

		// globalfee
		"/gaia.globalfee.v1beta1.Query/Params": func() proto.Message { return &globalfeetypes.QueryParamsResponse{} },

		// distribution
		"/cosmos.distribution.v1beta1.Query/DelegationRewards": func() proto.Message { return &types.QueryDelegationRewardsResponse{} },

		// staking
		"/cosmos.staking.v1beta1.Query/Delegation":                    func() proto.Message { return &stakingtypes.QueryDelegationResponse{} },
		"/cosmos.staking.v1beta1.Query/UnbondingDelegation":           func() proto.Message { return &stakingtypes.QueryUnbondingDelegationResponse{} },
		"/cosmos.staking.v1beta1.Query/Validator":                     func() proto.Message { return &stakingtypes.QueryValidatorResponse{} },
		"/cosmos.staking.v1beta1.Query/DelegatorDelegations":          func() proto.Message { return &stakingtypes.QueryDelegatorDelegationsResponse{} },
		"/cosmos.staking.v1beta1.Query/DelegatorUnbondingDelegations": func() proto.Message { return &stakingtypes.QueryDelegatorUnbondingDelegationsResponse{} },

		// harpoon
		"/neutron.harpoon.Query/SubscribedContracts": func() proto.Message { return &harpoontypes.QuerySubscribedContractsResponse{} },

		// state verifier
		"/neutron.state_verifier.v1.Query/VerifyStateValues": func() proto.Message { return &stateverifiertypes.QueryVerifyStateValuesResponse{} },
	}
}
