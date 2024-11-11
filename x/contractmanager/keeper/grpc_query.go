package keeper

import (
	"github.com/neutron-org/neutron/v5/x/contractmanager/types"
)

var _ types.QueryServer = Keeper{}

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}
