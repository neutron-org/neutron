package bindings

import (
	"encoding/json"

	"cosmossdk.io/math"

	contractmanagertypes "github.com/neutron-org/neutron/v3/x/contractmanager/types"
	dextypes "github.com/neutron-org/neutron/v3/x/dex/types"

	feerefundertypes "github.com/neutron-org/neutron/v3/x/feerefunder/types"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" // linter:staticcheck

	"github.com/neutron-org/neutron/v3/x/interchainqueries/types"
)

// NeutronQuery contains neutron custom queries.
type NeutronQuery struct {
	// Registered Interchain Query Result for specified QueryID
	InterchainQueryResult *QueryRegisteredQueryResultRequest `json:"interchain_query_result,omitempty"`
	// Interchain account address for specified ConnectionID and OwnerAddress
	InterchainAccountAddress *QueryInterchainAccountAddressRequest `json:"interchain_account_address,omitempty"`
	// RegisteredInterchainQueries
	RegisteredInterchainQueries *QueryRegisteredQueriesRequest `json:"registered_interchain_queries,omitempty"`
	// RegisteredInterchainQuery
	RegisteredInterchainQuery *QueryRegisteredQueryRequest `json:"registered_interchain_query,omitempty"`
	// TotalBurnedNeutronsAmount
	TotalBurnedNeutronsAmount *QueryTotalBurnedNeutronsAmountRequest `json:"total_burned_neutrons_amount,omitempty"`
	// MinIbcFee
	MinIbcFee *QueryMinIbcFeeRequest `json:"min_ibc_fee,omitempty"`
	// Token Factory queries
	// Given a subdenom minted by a contract via `NeutronMsg::MintTokens`,
	// returns the full denom as used by `BankMsg::Send`.
	FullDenom *FullDenom `json:"full_denom,omitempty"`
	// Returns the admin of a denom, if the denom is a Token Factory denom.
	DenomAdmin *DenomAdmin `json:"denom_admin,omitempty"`
	// Returns the before send hook if it was set before
	BeforeSendHook *BeforeSendHook `json:"before_send_hook,omitempty"`
	// Contractmanager queries
	// Query all failures for address
	Failures *Failures `json:"failures,omitempty"`
	// dex module queries
	Dex *DexQuery `json:"dex,omitempty"`
}

/* Requests */

type QueryRegisteredQueryResultRequest struct {
	QueryID uint64 `json:"query_id,omitempty"`
}

type QueryInterchainAccountAddressRequest struct {
	// owner_address is the owner of the interchain account on the controller chain
	OwnerAddress string `json:"owner_address,omitempty"`
	// interchain_account_id is an identifier of your interchain account from which you want to execute msgs
	InterchainAccountID string `json:"interchain_account_id,omitempty"`
	// connection_id is an IBC connection identifier between Neutron and remote chain
	ConnectionID string `json:"connection_id,omitempty"`
}

type QueryRegisteredQueriesRequest struct {
	Owners       []string           `json:"owners,omitempty"`
	ConnectionID string             `json:"connection_id,omitempty"`
	Pagination   *query.PageRequest `json:"pagination,omitempty"`
}

type QueryRegisteredQueryRequest struct {
	QueryID uint64 `json:"query_id,omitempty"`
}

/* Responses */

type QueryRegisteredQueryResponse struct {
	RegisteredQuery *RegisteredQuery `json:"registered_query,omitempty"`
}

type QueryRegisteredQueriesResponse struct {
	RegisteredQueries []RegisteredQuery `json:"registered_queries"`
}

type RegisteredQuery struct {
	// The unique id of the registered query.
	ID uint64 `json:"id"`
	// The address that registered the query.
	Owner string `json:"owner"`
	// The KV-storage keys for which we want to get values from remote chain
	Keys []*types.KVKey `json:"keys"`
	// The filter for transaction search ICQ
	TransactionsFilter string `json:"transactions_filter"`
	// The query type identifier (i.e. 'kv' or 'tx' for now).
	QueryType string `json:"query_type"`
	// The IBC connection ID for getting ConsensusState to verify proofs.
	ConnectionID string `json:"connection_id"`
	// Parameter that defines how often the query must be updated.
	UpdatePeriod uint64 `json:"update_period"`
	// The local chain last block height when the query result was updated.
	LastSubmittedResultLocalHeight uint64 `json:"last_submitted_result_local_height"`
	// The remote chain last block height when the query result was updated.
	LastSubmittedResultRemoteHeight *ibcclienttypes.Height `json:"last_submitted_result_remote_height,omitempty"`
	// Amount of coins deposited for the query.
	Deposit sdktypes.Coins `json:"deposit"`
	// Timeout before query becomes available for everybody to remove.
	SubmitTimeout uint64 `json:"submit_timeout"`
	// The local chain height when the query was registered.
	RegisteredAtHeight uint64 `json:"registered_at_height"`
}

type QueryTotalBurnedNeutronsAmountRequest struct{}

type QueryTotalBurnedNeutronsAmountResponse struct {
	Coin sdktypes.Coin `json:"coin"`
}

type QueryMinIbcFeeRequest struct{}

type QueryMinIbcFeeResponse struct {
	MinFee feerefundertypes.Fee `json:"min_fee"`
}

func (rq RegisteredQuery) MarshalJSON() ([]byte, error) {
	type AliasRQ RegisteredQuery

	a := struct {
		AliasRQ
	}{
		AliasRQ: (AliasRQ)(rq),
	}

	// We want keys be as empty array in Json ('[]'), not 'null'
	// It's easier to work with on smart-contracts side
	if a.Keys == nil {
		a.Keys = make([]*types.KVKey, 0)
	}

	return json.Marshal(a)
}

// Query response for an interchain account address
type QueryInterchainAccountAddressResponse struct {
	// The corresponding interchain account address on the host chain
	InterchainAccountAddress string `json:"interchain_account_address,omitempty"`
}

type QueryRegisteredQueryResultResponse struct {
	Result *QueryResult `json:"result,omitempty"`
}

type QueryResult struct {
	KvResults []*StorageValue `json:"kv_results,omitempty"`
	Height    uint64          `json:"height,omitempty"`
	Revision  uint64          `json:"revision,omitempty"`
}

type StorageValue struct {
	StoragePrefix string `json:"storage_prefix,omitempty"`
	Key           []byte `json:"key"`
	Value         []byte `json:"value"`
}

func (sv StorageValue) MarshalJSON() ([]byte, error) {
	type AliasSV StorageValue

	a := struct {
		AliasSV
	}{
		AliasSV: (AliasSV)(sv),
	}

	// We want Key and Value be as empty arrays in Json ('[]'), not 'null'
	// It's easier to work with on smart-contracts side
	if a.Key == nil {
		a.Key = make([]byte, 0)
	}
	if a.Value == nil {
		a.Value = make([]byte, 0)
	}

	return json.Marshal(a)
}

type FullDenom struct {
	CreatorAddr string `json:"creator_addr"`
	Subdenom    string `json:"subdenom"`
}

type DenomAdmin struct {
	Subdenom string `json:"subdenom"`
}

type BeforeSendHook struct {
	Denom string `json:"denom"`
}

type BeforeSendHookResponse struct {
	ContractAddr string `json:"contract_addr"`
}

type DenomAdminResponse struct {
	Admin string `json:"admin"`
}

type FullDenomResponse struct {
	Denom string `json:"denom"`
}

type Failures struct {
	Address    string             `json:"address"`
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

type FailuresResponse struct {
	Failures []contractmanagertypes.Failure `json:"failures"`
}

type DexQuery struct {
	// Parameters queries the parameters of the module.
	Params *dextypes.QueryParamsRequest `json:"params"`
	// Queries a LimitOrderTrancheUser by index.
	LimitOrderTrancheUser *dextypes.QueryGetLimitOrderTrancheUserRequest `json:"limit_order_tranche_user,omitempty"`
	// Queries a list of LimitOrderTrancheUser items.
	LimitOrderTrancheUserAll *dextypes.QueryAllLimitOrderTrancheUserRequest `json:"limit_order_tranche_user_all"`
	// Queries a list of LimitOrderTrancheUser items for a given address.
	LimitOrderTrancheUserAllByAddress *dextypes.QueryAllUserLimitOrdersRequest `json:"limit_order_tranche_user_all_by_address"`
	// Queries a LimitOrderTranche by index.
	LimitOrderTranche *dextypes.QueryGetLimitOrderTrancheRequest `json:"limit_order_tranche"`
	// Queries a list of LimitOrderTranche items for a given pairID / TokenIn combination.
	LimitOrderTrancheAll *dextypes.QueryAllLimitOrderTrancheRequest `json:"limit_order_tranche_all"`
	// Queries a list of UserDeposits items.
	UserDepositsAll *dextypes.QueryAllUserDepositsRequest `json:"user_deposit_all"`
	// Queries a list of TickLiquidity items.
	TickLiquidityAll *dextypes.QueryAllTickLiquidityRequest `json:"tick_liquidity_all"`
	// Queries a InactiveLimitOrderTranche by index.
	InactiveLimitOrderTranche *dextypes.QueryGetInactiveLimitOrderTrancheRequest `json:"inactive_limit_order_tranche"`
	// Queries a list of InactiveLimitOrderTranche items.
	InactiveLimitOrderTrancheAll *dextypes.QueryAllInactiveLimitOrderTrancheRequest `json:"inactive_limit_order_tranche_all"`
	// Queries a list of PoolReserves items.
	PoolReservesAll *dextypes.QueryAllPoolReservesRequest `json:"pool_reserves_all"`
	// Queries a PoolReserve by index
	PoolReserves *dextypes.QueryGetPoolReservesRequest `json:"pool_reserves"`
	// Queries the simulated result of a multihop swap
	EstimateMultiHopSwap *dextypes.QueryEstimateMultiHopSwapRequest `json:"estimate_multi_hop_swap"`
	// Queries the simulated result of a PlaceLimit order
	EstimatePlaceLimitOrder *QueryEstimatePlaceLimitOrderRequest `json:"estimate_place_limit_order"`
	// Queries a pool by pair, tick and fee
	Pool *dextypes.QueryPoolRequest `json:"pool"`
	// Queries a pool by ID
	PoolByID *dextypes.QueryPoolByIDRequest `json:"pool_by_id"`
	// Queries a PoolMetadata by ID
	PoolMetadata *dextypes.QueryGetPoolMetadataRequest `json:"pool_metadata"`
	// Queries a list of PoolMetadata items.
	PoolMetadataAll *dextypes.QueryAllPoolMetadataRequest `json:"pool_metadata_all"`
}

// QueryEstimatePlaceLimitOrderRequest is a copy dextypes.QueryEstimatePlaceLimitOrderRequest with altered ExpirationTime field,
// it's a preferable way to pass timestamp as unixtime to contracts
type QueryEstimatePlaceLimitOrderRequest struct {
	Creator          string   `json:"creator,omitempty"`
	Receiver         string   `json:"receiver,omitempty"`
	TokenIn          string   `json:"token_in,omitempty"`
	TokenOut         string   `json:"token_out,omitempty"`
	TickIndexInToOut int64    `json:"tick_index_in_to_out,omitempty"`
	AmountIn         math.Int `json:"amount_in"`
	OrderType        string   `json:"order_type,omitempty"`
	// expirationTime is only valid iff orderType == GOOD_TIL_TIME.
	ExpirationTime *uint64   `json:"expiration_time,omitempty"`
	MaxAmountOut   *math.Int `json:"max_amount_out"`
}
