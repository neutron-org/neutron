package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	"strings"
)

const delimeter = "."

type ICAOwner string

func (i ICAOwner) String() string {
	return string(i)
}

func NewICAOwner(contractAddress, owner string) ICAOwner {
	return ICAOwner(contractAddress + delimeter + owner)
}

func (i ICAOwner) GetContract() (sdk.AccAddress, error) {
	splittedOwner := strings.Split(strings.ReplaceAll(string(i), icatypes.PortPrefix, ""), delimeter)
	if len(splittedOwner) < 2 {
		return nil, sdkerrors.Wrap(ErrInvalidICAOwner, "invalid ICA owner format")
	}

	return sdk.AccAddressFromBech32(splittedOwner[0])
}
