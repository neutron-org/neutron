package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v5/x/state-verifier/types"
)

// VerifyStateValues implements `VerifyStateValues` gRPC query to verify storage values
func (k *Keeper) VerifyStateValues(ctx context.Context, request *types.QueryVerifyStateValuesRequest) (*types.QueryVerifyStateValuesResponse, error) {
	if err := k.Verify(sdk.UnwrapSDKContext(ctx), int64(request.Height), request.StorageValues); err != nil {
		return nil, err
	}

	return &types.QueryVerifyStateValuesResponse{Valid: true}, nil
}
