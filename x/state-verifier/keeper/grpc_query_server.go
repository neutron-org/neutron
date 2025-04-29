package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stateverifier "github.com/neutron-org/neutron/v6/x/state-verifier/types"
)

type queryServer struct {
	keeper *Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper *Keeper) stateverifier.QueryServer {
	return &queryServer{keeper: keeper}
}

// VerifyStateValues implements `VerifyStateValues` gRPC query to verify storage values
func (k *queryServer) VerifyStateValues(ctx context.Context, request *stateverifier.QueryVerifyStateValuesRequest) (*stateverifier.QueryVerifyStateValuesResponse, error) {
	if err := k.keeper.Verify(sdk.UnwrapSDKContext(ctx), int64(request.Height), request.StorageValues); err != nil { //nolint:gosec
		return nil, err
	}

	return &stateverifier.QueryVerifyStateValuesResponse{Valid: true}, nil
}
