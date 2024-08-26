package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v4/x/lastlook/types"
)

func (k Keeper) Batch(sdkCtx context.Context, request *types.QueryBatchRequest) (*types.QueryBatchResponse, error) {
	ctx := sdk.UnwrapSDKContext(sdkCtx)

	height := request.Height
	if height == 0 {
		height = ctx.BlockHeight()
	}

	batch, err := k.GetBatch(ctx, height)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get batch")
	}

	return &types.QueryBatchResponse{Batch: batch}, nil
}
