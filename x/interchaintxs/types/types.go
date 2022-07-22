package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	"strings"
)

const Delimiter = "."

type ICAOwner struct {
	contractAddress     sdk.AccAddress
	interchainAccountID string
}

func (i ICAOwner) String() string {
	return i.contractAddress.String() + Delimiter + i.interchainAccountID
}

func NewICAOwner(contractAddress, interchainAccountID string) (ICAOwner, error) {
	sdkContractAddress, err := sdk.AccAddressFromBech32(contractAddress)
	if err != nil {
		return ICAOwner{}, sdkerrors.Wrapf(ErrInvalidAccountAddress, "failed to decode address from bech32: %v", err)
	}
	return ICAOwner{contractAddress: sdkContractAddress, interchainAccountID: interchainAccountID}, nil
}

func ICAOwnerFromPort(port string) (ICAOwner, error) {
	splittedOwner := strings.SplitN(strings.TrimPrefix(port, icatypes.PortPrefix), Delimiter, 2)
	if len(splittedOwner) < 2 {
		return ICAOwner{}, sdkerrors.Wrap(ErrInvalidICAOwner, "invalid ICA interchainAccountID format")
	}

	contractAddress, err := sdk.AccAddressFromBech32(splittedOwner[0])
	if err != nil {
		return ICAOwner{}, sdkerrors.Wrapf(ErrInvalidAccountAddress, "failed to decode address from bech32: %v", err)
	}

	return ICAOwner{contractAddress: contractAddress, interchainAccountID: splittedOwner[1]}, nil
}

func (i ICAOwner) GetContract() sdk.AccAddress {
	return i.contractAddress
}

func (i ICAOwner) GetInterchainAccountID() string {
	return i.interchainAccountID
}
