package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	contractmanagertypes "github.com/neutron-org/neutron/x/contractmanager/types"
)

// ContractManagerKeeper defines the expected interface needed to add ack information about sudo failure.
type ContractManagerKeeper interface {
	AddContractFailure(ctx sdk.Context, failure contractmanagertypes.Failure)
}
