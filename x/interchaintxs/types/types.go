package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	"strings"
)

const delimiter = "."

type ICAOwner struct {
	contractAddress     sdk.AccAddress
	interchainAccountID string
}

func (i ICAOwner) String() string {
	return i.contractAddress.String() + delimiter + i.interchainAccountID
}

func NewICAOwner(contractAddress, interchainAccountID string) (ICAOwner, error) {
	// this is production version of the code
	// must be uncommented when the contracts are ready to send IC txs
	//
	//sdkContractAddress, err := sdk.AccAddressFromBech32(contractAddress)
	//if err != nil {
	//	return ICAOwner{}, sdkerrors.Wrapf(ErrInvalidAccountAddress, "failed to decode address from bech32: %v", err)
	//}
	//return ICAOwner{contractAddress: sdkContractAddress, interchainAccountID: interchainAccountID}, nil

	// this is ONLY for the demo scripts to see that Sudo actually works
	// this means anyone can set contractAddress
	sdkContractAddress, err := sdk.AccAddressFromBech32(interchainAccountID)
	if err != nil {
		return ICAOwner{}, sdkerrors.Wrapf(ErrInvalidAccountAddress, "failed to decode address from bech32: %v", err)
	}
	return ICAOwner{contractAddress: sdkContractAddress}, nil
}

func ICAOwnerFromPort(port string) (ICAOwner, error) {
	splittedOwner := strings.Split(strings.TrimPrefix(port, icatypes.PortPrefix), delimiter)
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
