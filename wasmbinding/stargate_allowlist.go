package wasmbinding

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v4/modules/core/03-connection/types"
	tokenfactorytypes "github.com/neutron-org/neutron/x/tokenfactory/types"
)

func AcceptedStargateQueries() wasmkeeper.AcceptedStargateQueries {
	// ===== JUNO
	return wasmkeeper.AcceptedStargateQueries{
		// ibc
		"/ibc.core.client.v1.Query/ClientState":    &ibcclienttypes.QueryClientStateResponse{},
		"/ibc.core.client.v1.Query/ConsensusState": &ibcclienttypes.QueryConsensusStateResponse{},
		"/ibc.core.connection.v1.Query/Connection": &ibcconnectiontypes.QueryConnectionResponse{},

		// governance
		//"/cosmos.gov.v1beta1.Query/Vote": &govv1.QueryVoteResponse{},

		// distribution
		//"/cosmos.distribution.v1beta1.Query/DelegationRewards": &distrtypes.QueryDelegationRewardsResponse{},

		// staking
		//"/cosmos.staking.v1beta1.Query/Delegation":          &stakingtypes.QueryDelegationResponse{},
		//"/cosmos.staking.v1beta1.Query/Redelegations":       &stakingtypes.QueryRedelegationsResponse{},
		//"/cosmos.staking.v1beta1.Query/UnbondingDelegation": &stakingtypes.QueryUnbondingDelegationResponse{},
		//"/cosmos.staking.v1beta1.Query/Validator":           &stakingtypes.QueryValidatorResponse{},
		//"/cosmos.staking.v1beta1.Query/Params":              &stakingtypes.QueryParamsResponse{},
		//"/cosmos.staking.v1beta1.Query/Pool":                &stakingtypes.QueryPoolResponse{},

		// token factory
		"/osmosis.tokenfactory.v1beta1.Query/Params":                 &tokenfactorytypes.QueryParamsResponse{},
		"/osmosis.tokenfactory.v1beta1.Query/DenomAuthorityMetadata": &tokenfactorytypes.QueryDenomAuthorityMetadataResponse{},
		"/osmosis.tokenfactory.v1beta1.Query/DenomsFromCreator":      &tokenfactorytypes.QueryDenomsFromCreatorResponse{},

		// ===== OSMOSIS
		// ibc queries
		"/ibc.applications.transfer.v1.Query/DenomTrace": &ibctransfertypes.QueryDenomTraceResponse{},

		// cosmos-sdk queries

		// auth
		"/cosmos.auth.v1beta1.Query/Account": &authtypes.QueryAccountResponse{},
		"/cosmos.auth.v1beta1.Query/Params":  &authtypes.QueryParamsResponse{},

		// bank
		"/cosmos.bank.v1beta1.Query/Balance":       &banktypes.QueryBalanceResponse{},
		"/cosmos.bank.v1beta1.Query/DenomMetadata": &banktypes.QueryDenomsMetadataResponse{},
		"/cosmos.bank.v1beta1.Query/Params":        &banktypes.QueryParamsResponse{},
		"/cosmos.bank.v1beta1.Query/SupplyOf":      &banktypes.QuerySupplyOfResponse{},

		// distribution
		//TODO: check what distribution module does. We disabled it. why?
		//"/cosmos.distribution.v1beta1.Query/Params":                   distributiontypes.QueryParamsResponse{},
		//"/cosmos.distribution.v1beta1.Query/DelegatorWithdrawAddress": distributiontypes.QueryDelegatorWithdrawAddressResponse{},
		//"/cosmos.distribution.v1beta1.Query/ValidatorCommission":      distributiontypes.QueryValidatorCommissionResponse{},

		// gov
		//"/cosmos.gov.v1beta1.Query/Deposit": govtypes.QueryDepositResponse{},
		//"/cosmos.gov.v1beta1.Query/Params":  govtypes.QueryParamsResponse{},
		//"/cosmos.gov.v1beta1.Query/Vote":    govtypes.QueryVoteResponse{},

		// slashing
		//TODO: no sense since provider only slashing?
		//"/cosmos.slashing.v1beta1.Query/Params":      slashingtypes.QueryParamsResponse{},
		//"/cosmos.slashing.v1beta1.Query/SigningInfo": slashingtypes.QuerySigningInfoResponse{},

		// staking
		//"/cosmos.staking.v1beta1.Query/Delegation": stakingtypes.QueryDelegationResponse{},
		//"/cosmos.staking.v1beta1.Query/Params":     stakingtypes.QueryParamsResponse{},
		//"/cosmos.staking.v1beta1.Query/Validator":  stakingtypes.QueryValidatorResponse{},

		// osmosis queries

		// epochs
		//"/osmosis.epochs.v1beta1.Query/EpochInfos":   epochtypes.QueryEpochsInfoResponse{},
		//"/osmosis.epochs.v1beta1.Query/CurrentEpoch": epochtypes.QueryCurrentEpochResponse{},

		// gamm
		//"/osmosis.gamm.v1beta1.Query/NumPools":                    gammtypes.QueryNumPoolsResponse{},
		//"/osmosis.gamm.v1beta1.Query/TotalLiquidity":              gammtypes.QueryTotalLiquidityResponse{},
		//"/osmosis.gamm.v1beta1.Query/Pool":                        gammtypes.QueryPoolResponse{},
		//"/osmosis.gamm.v1beta1.Query/PoolParams":                  gammtypes.QueryPoolParamsResponse{},
		//"/osmosis.gamm.v1beta1.Query/TotalPoolLiquidity":          gammtypes.QueryTotalPoolLiquidityResponse{},
		//"/osmosis.gamm.v1beta1.Query/TotalShares":                 gammtypes.QueryTotalSharesResponse{},
		//"/osmosis.gamm.v1beta1.Query/CalcJoinPoolShares":          gammtypes.QueryCalcJoinPoolSharesResponse{},
		//"/osmosis.gamm.v1beta1.Query/CalcExitPoolCoinsFromShares": gammtypes.QueryCalcExitPoolCoinsFromSharesResponse{},
		//"/osmosis.gamm.v1beta1.Query/CalcJoinPoolNoSwapShares":    gammtypes.QueryCalcJoinPoolNoSwapSharesResponse{},
		//"/osmosis.gamm.v1beta1.Query/PoolType":                    gammtypes.QueryPoolTypeResponse{},
		//"/osmosis.gamm.v2.Query/SpotPrice":                        gammv2types.QuerySpotPriceResponse{},
		//"/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountIn":   gammtypes.QuerySwapExactAmountInResponse{},
		//"/osmosis.gamm.v1beta1.Query/EstimateSwapExactAmountOut":  gammtypes.QuerySwapExactAmountOutResponse{},

		// incentives
		//"/osmosis.incentives.Query/ModuleToDistributeCoins": incentivestypes.ModuleToDistributeCoinsResponse{},
		//"/osmosis.incentives.Query/LockableDurations":       incentivestypes.QueryLockableDurationsResponse{},

		// lockup
		//"/osmosis.lockup.Query/ModuleBalance":          lockuptypes.ModuleBalanceResponse{},
		//"/osmosis.lockup.Query/ModuleLockedAmount":     lockuptypes.ModuleLockedAmountResponse{},
		//"/osmosis.lockup.Query/AccountUnlockableCoins": lockuptypes.AccountUnlockableCoinsResponse{},
		//"/osmosis.lockup.Query/AccountUnlockingCoins":  lockuptypes.AccountUnlockingCoinsResponse{},
		//"/osmosis.lockup.Query/LockedDenom":            lockuptypes.LockedDenomResponse{},
		//"/osmosis.lockup.Query/LockedByID":             lockuptypes.LockedResponse{},
		//"/osmosis.lockup.Query/NextLockID":             lockuptypes.NextLockIDResponse{},
		//"/osmosis.lockup.Query/LockRewardReceiver":     lockuptypes.LockRewardReceiverResponse{},

		// mint
		//"/osmosis.mint.v1beta1.Query/EpochProvisions": minttypes.QueryEpochProvisionsResponse{},
		//"/osmosis.mint.v1beta1.Query/Params":          minttypes.QueryParamsResponse{},

		// pool-incentives
		//"/osmosis.poolincentives.v1beta1.Query/GaugeIds": poolincentivestypes.QueryGaugeIdsResponse{},

		// superfluid
		//"/osmosis.superfluid.Query/Params":          superfluidtypes.QueryParamsResponse{},
		//"/osmosis.superfluid.Query/AssetType":       superfluidtypes.AssetTypeResponse{},
		//"/osmosis.superfluid.Query/AllAssets":       superfluidtypes.AllAssetsResponse{},
		//"/osmosis.superfluid.Query/AssetMultiplier": superfluidtypes.AssetMultiplierResponse{},

		// poolmanager
		//"/osmosis.poolmanager.v1beta1.Query/NumPools":                             poolmanagerqueryproto.NumPoolsResponse{},
		//"/osmosis.poolmanager.v1beta1.Query/EstimateSwapExactAmountIn":            poolmanagerqueryproto.EstimateSwapExactAmountInResponse{},
		//"/osmosis.poolmanager.v1beta1.Query/EstimateSwapExactAmountOut":           poolmanagerqueryproto.EstimateSwapExactAmountOutResponse{},
		//"/osmosis.poolmanager.v1beta1.Query/EstimateSinglePoolSwapExactAmountIn":  poolmanagerqueryproto.EstimateSwapExactAmountInResponse{},
		//"/osmosis.poolmanager.v1beta1.Query/EstimateSinglePoolSwapExactAmountOut": poolmanagerqueryproto.EstimateSwapExactAmountOutResponse{},
		//"/osmosis.poolmanager.v1beta1.Query/Pool":                                 poolmanagerqueryproto.PoolResponse{},
		//"/osmosis.poolmanager.v1beta1.Query/SpotPrice":                            poolmanagerqueryproto.SpotPriceResponse{},

		// txfees
		//"/osmosis.txfees.v1beta1.Query/FeeTokens":      txfeestypes.QueryFeeTokensResponse{},
		//"/osmosis.txfees.v1beta1.Query/DenomSpotPrice": txfeestypes.QueryDenomSpotPriceResponse{},
		//"/osmosis.txfees.v1beta1.Query/DenomPoolId":    txfeestypes.QueryDenomPoolIdResponse{},
		//"/osmosis.txfees.v1beta1.Query/BaseDenom":      txfeestypes.QueryBaseDenomResponse{},

		// twap
		//"/osmosis.twap.v1beta1.Query/ArithmeticTwap":      twapquerytypes.ArithmeticTwapResponse{},
		//"/osmosis.twap.v1beta1.Query/ArithmeticTwapToNow": twapquerytypes.ArithmeticTwapToNowResponse{},
		//"/osmosis.twap.v1beta1.Query/GeometricTwap":       twapquerytypes.GeometricTwapResponse{},
		//"/osmosis.twap.v1beta1.Query/GeometricTwapToNow":  twapquerytypes.GeometricTwapToNowResponse{},
		//"/osmosis.twap.v1beta1.Query/Params":              twapquerytypes.ParamsResponse{},

		// downtime-detector
		//"/osmosis.downtimedetector.v1beta1.Query/RecoveredSinceDowntimeOfLength": downtimequerytypes.RecoveredSinceDowntimeOfLengthResponse{},

		// concentrated-liquidity
		//"/osmosis.concentratedliquidity.v1beta1.Query/Pools":                                concentratedliquidityquery.PoolsResponse{},
		//"/osmosis.concentratedliquidity.v1beta1.Query/UserPositions":                        concentratedliquidityquery.UserPositionsResponse{},
		//"/osmosis.concentratedliquidity.v1beta1.Query/LiquidityPerTickRange":                concentratedliquidityquery.LiquidityPerTickRangeResponse{},
		//"/osmosis.concentratedliquidity.v1beta1.Query/LiquidityNetInDirection":              concentratedliquidityquery.LiquidityNetInDirectionResponse{},
		//"/osmosis.concentratedliquidity.v1beta1.Query/ClaimableSpreadRewards":               concentratedliquidityquery.ClaimableSpreadRewardsResponse{},
		//"/osmosis.concentratedliquidity.v1beta1.Query/ClaimableIncentives":                  concentratedliquidityquery.ClaimableIncentivesResponse{},
		//"/osmosis.concentratedliquidity.v1beta1.Query/PositionById":                         concentratedliquidityquery.PositionByIdResponse{},
		//"/osmosis.concentratedliquidity.v1beta1.Query/Params":                               concentratedliquidityquery.ParamsResponse{},
		//"/osmosis.concentratedliquidity.v1beta1.Query/PoolAccumulatorRewards":               concentratedliquidityquery.PoolAccumulatorRewardsResponse{},
		//"/osmosis.concentratedliquidity.v1beta1.Query/IncentiveRecords":                     concentratedliquidityquery.IncentiveRecordsResponse{},
		//"/osmosis.concentratedliquidity.v1beta1.Query/TickAccumulatorTrackers":              concentratedliquidityquery.TickAccumulatorTrackersResponse{},
		//"/osmosis.concentratedliquidity.v1beta1.Query/CFMMPoolIdLinkFromConcentratedPoolId": concentratedliquidityquery.CFMMPoolIdLinkFromConcentratedPoolIdResponse{},
	}
}
