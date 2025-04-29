package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	stateverifier "github.com/neutron-org/neutron/v7/x/state-verifier/types"
)

type queryServer struct {
	keeper *Keeper
}

// QueryConsensusState implements `QueryConsensusState` gRPC query to query storage values
func (q *queryServer) QueryConsensusState(ctx context.Context, request *stateverifier.QueryConsensusStateRequest) (*stateverifier.QueryConsensusStateResponse, error) {
	cs, err := q.keeper.GetConsensusState(sdk.UnwrapSDKContext(ctx), int64(request.Height))
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrKeyNotFound, "failed to get consensus state for height %d: %v", request.Height, err)
	}

	return &stateverifier.QueryConsensusStateResponse{Cs: cs}, nil
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper *Keeper) stateverifier.QueryServer {
	return &queryServer{keeper: keeper}
}

// VerifyStateValues implements `VerifyStateValues` gRPC query to verify storage values
func (q *queryServer) VerifyStateValues(ctx context.Context, request *stateverifier.QueryVerifyStateValuesRequest) (*stateverifier.QueryVerifyStateValuesResponse, error) {
	if err := q.keeper.Verify(sdk.UnwrapSDKContext(ctx), int64(request.Height), request.StorageValues); err != nil { //nolint:gosec
		return nil, err
	}

	return &stateverifier.QueryVerifyStateValuesResponse{Valid: true}, nil
}
