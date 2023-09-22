package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/utils"
	"github.com/neutron-org/neutron/x/dex/types"
	dexutils "github.com/neutron-org/neutron/x/dex/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) UserDepositsAll(
	goCtx context.Context,
	req *types.QueryAllUserDepositsRequest,
) (*types.QueryAllUserDepositsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	var depositArr []*types.DepositRecord

	pageRes, err := utils.FilteredPaginateAccountBalances(
		ctx,
		k.bankKeeper,
		addr,
		req.Pagination,
		func(poolCoinMaybe sdk.Coin, accumulate bool) bool {
			err := types.ValidatePoolDenom(poolCoinMaybe.Denom)
			if err != nil {
				return false
			}

			poolMetadata, err := k.GetPoolMetadataByDenom(ctx, poolCoinMaybe.Denom)
			if err != nil {
				panic("Can't get info for PoolDenom")
			}

			fee := dexutils.MustSafeUint64ToInt64(poolMetadata.Fee)

			if accumulate {
				depositRecord := &types.DepositRecord{
					PairID:          poolMetadata.PairID,
					SharesOwned:     poolCoinMaybe.Amount,
					CenterTickIndex: poolMetadata.Tick,
					LowerTickIndex:  poolMetadata.Tick - fee,
					UpperTickIndex:  poolMetadata.Tick + fee,
					Fee:             poolMetadata.Fee,
				}

				depositArr = append(depositArr, depositRecord)
			}

			return true
		})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllUserDepositsResponse{
		Deposits:   depositArr,
		Pagination: pageRes,
	}, nil
}
