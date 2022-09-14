package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (queryInfo *RegisteredQuery) GetOwnerAddress() (creator sdk.AccAddress, err error) {
	creator, err = sdk.AccAddressFromBech32(queryInfo.Owner)
	if err != nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to decode owner address: %s", queryInfo.Owner)
	}

	return creator, nil
}
