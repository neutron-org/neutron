package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/neutron-org/neutron/v6/utils"
	"github.com/neutron-org/neutron/v6/x/dex/types"
	dexutils "github.com/neutron-org/neutron/v6/x/dex/utils"
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
					PairId:          poolMetadata.PairId,
					SharesOwned:     poolCoinMaybe.Amount,
					CenterTickIndex: poolMetadata.Tick,
					LowerTickIndex:  poolMetadata.Tick - fee,
					UpperTickIndex:  poolMetadata.Tick + fee,
					Fee:             poolMetadata.Fee,
				}

				if req.IncludePoolData {
					k.addPoolData(ctx, depositRecord)
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

func (k Keeper) addPoolData(ctx sdk.Context, record *types.DepositRecord) *types.DepositRecord {
	pool, found := k.GetPool(ctx, record.PairId, record.CenterTickIndex, record.Fee)
	if !found {
		panic("Pool does not exist")
	}

	record.Pool = pool
	supply := k.bankKeeper.GetSupply(ctx, pool.GetPoolDenom())
	record.TotalShares = &supply.Amount

	return record
}
