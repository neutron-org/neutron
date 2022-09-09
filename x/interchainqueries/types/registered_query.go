package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (queryInfo *RegisteredQuery) GetOwnerAddress() (creator sdk.AccAddress, err error) {
	creator, errCreator := sdk.AccAddressFromBech32(queryInfo.Owner)
	return creator, sdkerrors.Wrapf(errCreator, ErrInvalidOwner.Error(), queryInfo.Owner)
}

func (queryInfo *RegisteredQuery) GetDepositCoin() (wager sdk.Coin) {
	return sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(int64(queryInfo.Deposit)))
}
