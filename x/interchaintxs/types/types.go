package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/types"
)

const Delimiter = "."

type ICAOwner struct {
	contractAddress     sdk.AccAddress
	interchainAccountID string
}

func (i ICAOwner) String() string {
	return i.contractAddress.String() + Delimiter + i.interchainAccountID
}

func NewICAOwner(contractAddressBech32, interchainAccountID string) (ICAOwner, error) {
	sdkContractAddress, err := sdk.AccAddressFromBech32(contractAddressBech32)
	if err != nil {
		return ICAOwner{}, sdkerrors.Wrapf(ErrInvalidAccountAddress, "failed to decode address from bech32: %v", err)
	}

	return ICAOwner{contractAddress: sdkContractAddress, interchainAccountID: interchainAccountID}, nil
}

func NewICAOwnerFromAddress(address sdk.AccAddress, interchainAccountID string) ICAOwner {
	return ICAOwner{contractAddress: address, interchainAccountID: interchainAccountID}
}

func ICAOwnerFromPort(port string) (ICAOwner, error) {
	splitOwner := strings.SplitN(strings.TrimPrefix(port, icatypes.PortPrefix), Delimiter, 2)
	if len(splitOwner) < 2 {
		return ICAOwner{}, sdkerrors.Wrap(ErrInvalidICAOwner, "invalid ICA interchainAccountID format")
	}

	contractAddress, err := sdk.AccAddressFromBech32(splitOwner[0])
	if err != nil {
		return ICAOwner{}, sdkerrors.Wrapf(ErrInvalidAccountAddress, "failed to decode address from bech32: %v", err)
	}

	return ICAOwner{contractAddress: contractAddress, interchainAccountID: splitOwner[1]}, nil
}

func (i ICAOwner) GetContract() sdk.AccAddress {
	return i.contractAddress
}

func (i ICAOwner) GetInterchainAccountID() string {
	return i.interchainAccountID
}
