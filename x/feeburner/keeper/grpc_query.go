package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/feeburner/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) TotalBurnedNeutronsAmount(goCtx context.Context, _ *types.QueryTotalBurnedNeutronsAmountRequest) (*types.QueryTotalBurnedNeutronsAmountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	totalBurnedNeutronsAmount := k.GetTotalBurnedNeutronsAmount(ctx)

	return &types.QueryTotalBurnedNeutronsAmountResponse{TotalBurnedNeutronsAmount: totalBurnedNeutronsAmount}, nil
}
