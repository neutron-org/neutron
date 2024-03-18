package types

import (
	"fmt"

	"cosmossdk.io/errors"

	// this line is used by starport scaffolding # genesis/types/import
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for _, info := range gs.FeeInfos {
		addr, err := sdk.AccAddressFromBech32(info.Payer)
		if err != nil {
			return fmt.Errorf("failed to parse the payer address %s: %w", info.Payer, err)
		}

		if len(addr) != wasmtypes.ContractAddrLen {
			return errors.Wrapf(sdkerrors.ErrInvalidAddress, "fee payer address %s is not a contract", info.Payer)
		}

		if err := host.PortIdentifierValidator(info.PacketId.PortId); err != nil {
			return fmt.Errorf("port id %s is invalid: %w", info.PacketId.PortId, err)
		}

		if err := host.ChannelIdentifierValidator(info.PacketId.ChannelId); err != nil {
			return fmt.Errorf("channel id %s is invalid: %w", info.PacketId.PortId, err)
		}

		if err := info.Fee.Validate(); err != nil {
			return fmt.Errorf("invalid fees %s: %w", info.Fee, err)
		}
	}
	return gs.Params.Validate()
}
