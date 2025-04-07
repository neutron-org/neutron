package keeper_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	neutronapp "github.com/neutron-org/neutron/v6/app"
	"github.com/neutron-org/neutron/v6/testutil"
	keepertest "github.com/neutron-org/neutron/v6/x/dex/keeper/internal/testutils"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func simulateDeposit(ctx sdk.Context, app *neutronapp.App, addr sdk.AccAddress, deposit *types.DepositRecord) *types.Pool {
	// NOTE: For simplicity sake, we are not actually doing a deposit, we are just initializing
	// the pool and adding the poolDenom to the users account
	pool, err := app.DexKeeper.InitPool(ctx, deposit.PairId, deposit.CenterTickIndex, deposit.Fee)
	if err != nil {
		panic("Cannot init pool")
	}
	coin := sdk.NewCoin(pool.GetPoolDenom(), deposit.SharesOwned)
	keepertest.FundAccount(app.BankKeeper, ctx, addr, sdk.Coins{coin})

	return pool
}

func TestUserDepositsAllQueryPaginated(t *testing.T) {
	app := testutil.Setup(t)
	keeper := app.(*neutronapp.App).DexKeeper
	// `NewUncachedContext` like a `NewContext` calls `sdk.NewContext` under the hood. But the reason why we switched to NewUncachedContext
	// is NewContext tries to pass `app.finalizeBlockState.ms` as first argument while  app.finalizeBlockState is nil at this stage,
	// and we get nil pointer exception
	// when NewUncachedContext passes `app.cms` (multistore) as an argument to `sdk.NewContext`
	ctx := app.(*neutronapp.App).BaseApp.NewUncachedContext(false, tmproto.Header{})
	addr := sdk.AccAddress("test_addr")
	msgs := []*types.DepositRecord{
		{
			PairId:          defaultPairID,
			SharesOwned:     math.NewInt(10),
			CenterTickIndex: 2,
			LowerTickIndex:  1,
			UpperTickIndex:  3,
			Fee:             1,
		},
		{
			PairId:          defaultPairID,
			SharesOwned:     math.NewInt(10),
			CenterTickIndex: 3,
			LowerTickIndex:  2,
			UpperTickIndex:  4,
			Fee:             1,
		},
		{
			PairId:          defaultPairID,
			SharesOwned:     math.NewInt(10),
			CenterTickIndex: 4,
			LowerTickIndex:  3,
			UpperTickIndex:  5,
			Fee:             1,
		},
		{
			PairId:          defaultPairID,
			SharesOwned:     math.NewInt(10),
			CenterTickIndex: 5,
			LowerTickIndex:  4,
			UpperTickIndex:  6,
			Fee:             1,
		},
		{
			PairId:          defaultPairID,
			SharesOwned:     math.NewInt(10),
			CenterTickIndex: 6,
			LowerTickIndex:  5,
			UpperTickIndex:  7,
			Fee:             1,
		},
	}
	poolsFromDeposits := make([]*types.Pool, 0)
	for _, d := range msgs {
		pool := simulateDeposit(ctx, app.(*neutronapp.App), addr, d)
		poolsFromDeposits = append(poolsFromDeposits, pool)

	}

	// Add random noise to make sure only pool denoms are picked up
	randomCoins := sdk.Coins{sdk.NewInt64Coin("TokenA", 10), sdk.NewInt64Coin("TokenZ", 10)}
	keepertest.FundAccount(app.(*neutronapp.App).BankKeeper, ctx, addr, randomCoins)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllUserDepositsRequest {
		return &types.QueryAllUserDepositsRequest{
			Address: addr.String(),
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.UserDepositsAll(ctx, request(nil, uint64(i), uint64(step), false)) //nolint:gosec
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Deposits), step)
			require.Subset(t,
				msgs,
				resp.Deposits,
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		var allRecords []*types.DepositRecord
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.UserDepositsAll(ctx, request(next, 0, uint64(step), false)) //nolint:gosec
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Deposits), step)
			require.Subset(t,
				msgs,
				resp.Deposits,
			)

			allRecords = append(allRecords, resp.Deposits...)
			next = resp.Pagination.NextKey
		}
		require.ElementsMatch(t,
			msgs,
			allRecords,
		)
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.UserDepositsAll(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, uint64(len(msgs)), resp.Pagination.Total)
		require.ElementsMatch(t,
			msgs,
			resp.Deposits,
		)
	})
	t.Run("WithPoolData", func(t *testing.T) {
		req := request(nil, 0, 2, false)
		req.IncludePoolData = true
		resp, err := keeper.UserDepositsAll(ctx, req)
		require.NoError(t, err)
		require.Equal(t, len(resp.Deposits), 2)
		expectedShares := math.NewInt(10)
		expectedRespWithPoolData := []*types.DepositRecord{
			{
				PairId:          defaultPairID,
				SharesOwned:     math.NewInt(10),
				CenterTickIndex: 2,
				LowerTickIndex:  1,
				UpperTickIndex:  3,
				Fee:             1,
				Pool:            poolsFromDeposits[0],
				TotalShares:     &expectedShares,
			},
			{
				PairId:          defaultPairID,
				SharesOwned:     math.NewInt(10),
				CenterTickIndex: 3,
				LowerTickIndex:  2,
				UpperTickIndex:  4,
				Fee:             1,
				Pool:            poolsFromDeposits[1],
				TotalShares:     &expectedShares,
			},
		}
		require.ElementsMatch(t,
			expectedRespWithPoolData,
			resp.Deposits,
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.UserDepositsAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
