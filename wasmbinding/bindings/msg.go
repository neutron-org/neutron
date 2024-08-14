//nolint:revive,stylecheck  // if we change the names of var-naming things here, we harm some kind of mapping.
package bindings

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramChange "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	dextypes "github.com/neutron-org/neutron/v4/x/dex/types"
	feetypes "github.com/neutron-org/neutron/v4/x/feerefunder/types"
	icqtypes "github.com/neutron-org/neutron/v4/x/interchainqueries/types"
	transferwrappertypes "github.com/neutron-org/neutron/v4/x/transfer/types"
)

// ProtobufAny is a hack-struct to serialize protobuf Any message into JSON object
type ProtobufAny struct {
	TypeURL string `json:"type_url"`
	Value   []byte `json:"value"`
}

// NeutronMsg is used like a sum type to hold one of custom Neutron messages.
// Follow https://github.com/neutron-org/neutron-sdk/blob/main/packages/neutron-sdk/src/bindings/msg.rs
// for more information.
type NeutronMsg struct {
	SubmitTx                  *SubmitTx                         `json:"submit_tx,omitempty"`
	RegisterInterchainAccount *RegisterInterchainAccount        `json:"register_interchain_account,omitempty"`
	RegisterInterchainQuery   *RegisterInterchainQuery          `json:"register_interchain_query,omitempty"`
	UpdateInterchainQuery     *UpdateInterchainQuery            `json:"update_interchain_query,omitempty"`
	RemoveInterchainQuery     *RemoveInterchainQuery            `json:"remove_interchain_query,omitempty"`
	IBCTransfer               *transferwrappertypes.MsgTransfer `json:"ibc_transfer,omitempty"`
	SubmitAdminProposal       *SubmitAdminProposal              `json:"submit_admin_proposal,omitempty"`

	// Token factory types
	/// Contracts can create denoms, namespaced under the contract's address.
	/// A contract may create any number of independent sub-denoms.
	CreateDenom *CreateDenom `json:"create_denom,omitempty"`
	/// Contracts can change the admin of a denom that they are the admin of.
	ChangeAdmin *ChangeAdmin `json:"change_admin,omitempty"`
	/// Contracts can mint native tokens for an existing factory denom
	/// that they are the admin of.
	MintTokens *MintTokens `json:"mint_tokens,omitempty"`
	/// Contracts can burn native tokens for an existing factory denom
	/// that they are the admin of.
	/// Currently, the burn from address must be the admin contract.
	BurnTokens *BurnTokens `json:"burn_tokens,omitempty"`
	/// Contracts can set before send hook for an existing factory denom
	///	that they are the admin of.
	///	Currently, the set before hook call should be performed from address that must be the admin contract.
	SetBeforeSendHook *SetBeforeSendHook `json:"set_before_send_hook,omitempty"`
	/// Force transferring of a specific denom is only allowed for the creator of the denom registered during CreateDenom.
	ForceTransfer *ForceTransfer `json:"force_transfer,omitempty"`
	/// Setting of metadata for a specific denom is only allowed for the admin of the denom.
	/// It allows the overwriting of the denom metadata in the bank module.
	SetDenomMetadata *SetDenomMetadata `json:"set_denom_metadata,omitempty"`

	// Cron types
	AddSchedule    *AddSchedule    `json:"add_schedule,omitempty"`
	RemoveSchedule *RemoveSchedule `json:"remove_schedule,omitempty"`

	// Contractmanager types
	/// A contract that has failed acknowledgement can resubmit it
	ResubmitFailure *ResubmitFailure `json:"resubmit_failure,omitempty"`

	// dex module bindings
	Dex *Dex `json:"dex,omitempty"`
}

// SubmitTx submits interchain transaction on a remote chain.
type SubmitTx struct {
	ConnectionId        string        `json:"connection_id"`
	InterchainAccountId string        `json:"interchain_account_id"`
	Msgs                []ProtobufAny `json:"msgs"`
	Memo                string        `json:"memo"`
	Timeout             uint64        `json:"timeout"`
	Fee                 feetypes.Fee  `json:"fee"`
}

// RegisterInterchainAccount creates account on remote chain.
type RegisterInterchainAccount struct {
	ConnectionId        string    `json:"connection_id"`
	InterchainAccountId string    `json:"interchain_account_id"`
	RegisterFee         sdk.Coins `json:"register_fee,omitempty"`
}

// RegisterInterchainAccountResponse holds response for RegisterInterchainAccount.
type RegisterInterchainAccountResponse struct {
	ChannelId string `json:"channel_id"`
	PortId    string `json:"port_id"`
}

// RegisterInterchainQuery creates a query for remote chain.
type RegisterInterchainQuery struct {
	QueryType          string            `json:"query_type"`
	Keys               []*icqtypes.KVKey `json:"keys"`
	TransactionsFilter string            `json:"transactions_filter"`
	ConnectionId       string            `json:"connection_id"`
	UpdatePeriod       uint64            `json:"update_period"`
}

type SubmitAdminProposal struct {
	AdminProposal AdminProposal `json:"admin_proposal"`
}

type AdminProposal struct {
	ParamChangeProposal    *ParamChangeProposal    `json:"param_change_proposal,omitempty"`
	ProposalExecuteMessage *ProposalExecuteMessage `json:"proposal_execute_message,omitempty"`
}

type ParamChangeProposal struct {
	Title        string                    `json:"title"`
	Description  string                    `json:"description"`
	ParamChanges []paramChange.ParamChange `json:"param_changes"`
}

type Plan struct {
	Name   string `json:"name"`
	Height int64  `json:"height"`
	Info   string `json:"info"`
}

// RegisterInterchainQueryResponse holds response for RegisterInterchainQuery
type RegisterInterchainQueryResponse struct {
	Id uint64 `json:"id"`
}

type RemoveInterchainQuery struct {
	QueryId uint64 `json:"query_id"`
}

type RemoveInterchainQueryResponse struct{}

type UpdateInterchainQuery struct {
	QueryId               uint64            `json:"query_id,omitempty"`
	NewKeys               []*icqtypes.KVKey `json:"new_keys,omitempty"`
	NewUpdatePeriod       uint64            `json:"new_update_period,omitempty"`
	NewTransactionsFilter string            `json:"new_transactions_filter,omitempty"`
}

type UpdateInterchainQueryResponse struct{}

type ProposalExecuteMessage struct {
	Message string `json:"message,omitempty"`
}

// CreateDenom creates a new factory denom, of denomination:
// factory/{creating contract address}/{Subdenom}
// Subdenom can be of length at most 44 characters, in [0-9a-zA-Z./]
// The (creating contract address, subdenom) pair must be unique.
// The created denom's admin is the creating contract address,
// but this admin can be changed using the ChangeAdmin binding.
type CreateDenom struct {
	Subdenom string `json:"subdenom"`
}

// ChangeAdmin changes the admin for a factory denom.
// If the NewAdminAddress is empty, the denom has no admin.
type ChangeAdmin struct {
	Denom           string `json:"denom"`
	NewAdminAddress string `json:"new_admin_address"`
}

type MintTokens struct {
	Denom         string   `json:"denom"`
	Amount        math.Int `json:"amount"`
	MintToAddress string   `json:"mint_to_address"`
}

type BurnTokens struct {
	Denom  string   `json:"denom"`
	Amount math.Int `json:"amount"`
	// BurnFromAddress must be set to "" for now.
	BurnFromAddress string `json:"burn_from_address"`
}

// SetBeforeSendHook Allowing to assign a CosmWasm contract to call with a BeforeSend hook for a specific denom is only
// allowed for the creator of the denom registered during CreateDenom.
type SetBeforeSendHook struct {
	Denom        string `json:"denom"`
	ContractAddr string `json:"contract_addr"`
}

// SetDenomMetadata is sets the denom's bank metadata
type SetDenomMetadata struct {
	banktypes.Metadata
}

// ForceTransfer forces transferring of a specific denom is only allowed for the creator of the denom registered during CreateDenom.
type ForceTransfer struct {
	Denom               string   `json:"denom"`
	Amount              math.Int `json:"amount"`
	TransferFromAddress string   `json:"transfer_from_address"`
	TransferToAddress   string   `json:"transfer_to_address"`
}

// AddSchedule adds new schedule to the cron module
type AddSchedule struct {
	Name    string               `json:"name"`
	Period  uint64               `json:"period"`
	Msgs    []MsgExecuteContract `json:"msgs"`
	Blocker uint64               `json:"blocker"`
}

// AddScheduleResponse holds response AddSchedule
type AddScheduleResponse struct{}

// RemoveSchedule removes existing schedule with given name
type RemoveSchedule struct {
	Name string `json:"name"`
}

// RemoveScheduleResponse holds response RemoveSchedule
type RemoveScheduleResponse struct{}

// MsgExecuteContract defined separate from wasmtypes since we can get away with just passing the string into bindings
type MsgExecuteContract struct {
	// Contract is the address of the smart contract
	Contract string `json:"contract,omitempty"`
	// Msg json encoded message to be passed to the contract
	Msg string `json:"msg,omitempty"`
}

type ResubmitFailure struct {
	FailureId uint64 `json:"failure_id"`
}

type ResubmitFailureResponse struct {
	FailureId uint64 `json:"failure_id"`
}

type Dex struct {
	Deposit                  *dextypes.MsgDeposit                  `json:"deposit"`
	Withdrawal               *dextypes.MsgWithdrawal               `json:"withdrawal"`
	PlaceLimitOrder          *MsgPlaceLimitOrder                   `json:"place_limit_order"`
	WithdrawFilledLimitOrder *dextypes.MsgWithdrawFilledLimitOrder `json:"withdraw_filled_limit_order"`
	CancelLimitOrder         *dextypes.MsgCancelLimitOrder         `json:"cancel_limit_order"`
	MultiHopSwap             *dextypes.MsgMultiHopSwap             `json:"multi_hop_swap"`
}

// MsgPlaceLimitOrder is a copy dextypes.MsgPlaceLimitOrder with altered ExpirationTime field,
// it's a preferable way to pass timestamp as unixtime to contracts
type MsgPlaceLimitOrder struct {
	Creator  string `json:"creator,omitempty"`
	Receiver string `json:"receiver,omitempty"`
	TokenIn  string `json:"token_in,omitempty"`
	TokenOut string `json:"token_out,omitempty"`
	// Deprecated: tick_index_in_to_out will be removed in future release; limit_sell_price should be used instead.
	TickIndexInToOut int64    `json:"tick_index_in_to_out,omitempty"`
	AmountIn         math.Int `json:"amount_in"`
	OrderType        string   `json:"order_type,omitempty"`
	// expirationTime is only valid iff orderType == GOOD_TIL_TIME.
	ExpirationTime *uint64   `json:"expiration_time,omitempty"`
	MaxAmountOut   *math.Int `json:"max_amount_out"`
	// Accepts standard decimals and decimals with scientific notation (ie. 1234.23E-7)
	LimitSellPrice string `json:"limit_sell_price,omitempty"`
}
