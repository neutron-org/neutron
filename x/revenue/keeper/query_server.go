package keeper

import (
	"context"

	revenuetypes "github.com/neutron-org/neutron/v5/x/revenue/types"
)

type queryServer struct {
	keeper *Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper *Keeper) revenuetypes.QueryServer {
	return &queryServer{keeper: keeper}
}

var _ revenuetypes.QueryServer = queryServer{}

func (s queryServer) Params(ctx context.Context, request *revenuetypes.QueryParamsRequest) (*revenuetypes.QueryParamsResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (s queryServer) State(ctx context.Context, request *revenuetypes.QueryStateRequest) (*revenuetypes.QueryStateResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (s queryServer) ValidatorStats(ctx context.Context, request *revenuetypes.QueryValidatorStatsRequest) (*revenuetypes.QueryValidatorStatsResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (s queryServer) ValidatorsStats(ctx context.Context, request *revenuetypes.QueryValidatorsStatsRequest) (*revenuetypes.QueryValidatorsStatsResponse, error) {
	// TODO implement me
	panic("implement me")
}
