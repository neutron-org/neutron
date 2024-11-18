package types

import (
	"strings"

	"cosmossdk.io/errors"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg                            = &MsgSubmitQueryResult{}
	_ codectypes.UnpackInterfacesMessage = MsgSubmitQueryResult{}
)

func (msg MsgSubmitQueryResult) Route() string {
	return RouterKey
}

func (msg MsgSubmitQueryResult) Type() string {
	return "submit-query-result"
}

func (msg MsgSubmitQueryResult) Validate() error {
	if msg.Result == nil {
		return errors.Wrap(ErrEmptyResult, "query result can't be empty")
	}

	if len(msg.Result.KvResults) == 0 && msg.Result.Block == nil {
		return errors.Wrap(ErrEmptyResult, "query result can't be empty")
	}

	if msg.QueryId == 0 {
		return errors.Wrap(ErrInvalidQueryID, "query id cannot be equal zero")
	}

	if strings.TrimSpace(msg.Sender) == "" {
		return errors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse address: %s", msg.Sender)
	}

	return nil
}

func (msg MsgSubmitQueryResult) GetSignBytes() []byte {
	return ModuleCdc.MustMarshalJSON(&msg)
}

func (msg MsgSubmitQueryResult) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgSubmitQueryResult) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var header exported.ClientMessage
	if err := unpacker.UnpackAny(msg.Result.GetBlock().GetHeader(), &header); err != nil {
		return err
	}

	return unpacker.UnpackAny(msg.Result.GetBlock().GetNextBlockHeader(), &header)
}

//----------------------------------------------------------------

var _ sdk.Msg = &MsgRegisterInterchainQuery{}

func (msg MsgRegisterInterchainQuery) Route() string {
	return RouterKey
}

func (msg MsgRegisterInterchainQuery) Type() string {
	return "register-interchain-query"
}

func (msg MsgRegisterInterchainQuery) Validate(params Params) error {
	if msg.UpdatePeriod == 0 {
		return errors.Wrap(ErrInvalidUpdatePeriod, "update period can not be equal to zero")
	}

	if strings.TrimSpace(msg.Sender) == "" {
		return errors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse address: %s", msg.Sender)
	}

	if strings.TrimSpace(msg.ConnectionId) == "" {
		return errors.Wrap(ErrInvalidConnectionID, "connection id cannot be empty")
	}

	if !InterchainQueryType(msg.QueryType).IsValid() {
		return errors.Wrap(ErrInvalidQueryType, "invalid query type")
	}

	if InterchainQueryType(msg.QueryType).IsKV() {
		if len(msg.Keys) == 0 {
			return errors.Wrap(ErrEmptyKeys, "keys cannot be empty")
		}
		if err := validateKeys(msg.GetKeys(), params.MaxKvQueryKeysCount); err != nil {
			return err
		}
	}

	if InterchainQueryType(msg.QueryType).IsTX() {
		if err := ValidateTransactionsFilter(msg.TransactionsFilter, params.MaxTransactionsFilters); err != nil {
			return errors.Wrap(ErrInvalidTransactionsFilter, err.Error())
		}
	}
	return nil
}

func (msg MsgRegisterInterchainQuery) GetSignBytes() []byte {
	return ModuleCdc.MustMarshalJSON(&msg)
}

func (msg MsgRegisterInterchainQuery) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}

//----------------------------------------------------------------

var _ sdk.Msg = &MsgUpdateInterchainQueryRequest{}

func (msg MsgUpdateInterchainQueryRequest) Validate(params Params) error {
	if msg.GetQueryId() == 0 {
		return errors.Wrap(ErrInvalidQueryID, "query_id cannot be empty or equal to 0")
	}

	newKeys := msg.GetNewKeys()
	newTxFilter := msg.GetNewTransactionsFilter()

	if len(newKeys) == 0 && newTxFilter == "" && msg.GetNewUpdatePeriod() == 0 {
		return errors.Wrap(
			sdkerrors.ErrInvalidRequest,
			"one of new_keys, new_transactions_filter or new_update_period should be set",
		)
	}

	if len(newKeys) != 0 && newTxFilter != "" {
		return errors.Wrap(
			sdkerrors.ErrInvalidRequest,
			"either new_keys or new_transactions_filter should be set",
		)
	}

	if len(newKeys) != 0 {
		if err := validateKeys(newKeys, params.MaxKvQueryKeysCount); err != nil {
			return err
		}
	}

	if newTxFilter != "" {
		if err := ValidateTransactionsFilter(newTxFilter, params.MaxTransactionsFilters); err != nil {
			return errors.Wrap(ErrInvalidTransactionsFilter, err.Error())
		}
	}

	if strings.TrimSpace(msg.Sender) == "" {
		return errors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to parse address: %s", msg.Sender)
	}
	return nil
}

func (msg MsgUpdateInterchainQueryRequest) GetSignBytes() []byte {
	return ModuleCdc.MustMarshalJSON(&msg)
}

func (msg MsgUpdateInterchainQueryRequest) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}

//----------------------------------------------------------------

var _ sdk.Msg = &MsgUpdateParams{}

func (msg *MsgUpdateParams) Route() string {
	return RouterKey
}

func (msg *MsgUpdateParams) Type() string {
	return "update-params"
}

func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil { // should never happen as valid basic rejects invalid addresses
		panic(err.Error())
	}
	return []sdk.AccAddress{authority}
}

func (msg *MsgUpdateParams) GetSignBytes() []byte {
	return ModuleCdc.MustMarshalJSON(msg)
}

func (msg *MsgUpdateParams) Validate() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errors.Wrap(err, "authority is invalid")
	}
	return nil
}

func validateKeys(keys []*KVKey, maxKVQueryKeysCount uint64) error {
	if uint64(len(keys)) > maxKVQueryKeysCount {
		return errors.Wrapf(ErrTooManyKVQueryKeys, "keys count cannot be more than %d", maxKVQueryKeysCount)
	}

	duplicates := make(map[string]struct{})
	for _, key := range keys {
		if key == nil {
			return errors.Wrap(sdkerrors.ErrInvalidType, "key cannot be nil")
		}

		if _, ok := duplicates[key.ToString()]; ok {
			return errors.Wrap(sdkerrors.ErrInvalidRequest, "keys cannot be duplicated")
		}

		if len(key.Path) == 0 {
			return errors.Wrap(ErrEmptyKeyPath, "keys path cannot be empty")
		}

		if len(key.Key) == 0 {
			return errors.Wrap(ErrEmptyKeyID, "keys id cannot be empty")
		}

		duplicates[key.ToString()] = struct{}{}
	}

	return nil
}
