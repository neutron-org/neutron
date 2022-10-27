package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ContractManagerKeeper defines the expected interface needed to add ack information about sudo failure.
type ContractManagerKeeper interface {
	AddContractFailure(ctx sdk.Context, address string, ackId uint64, ackType string)
}
