package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/neutron-org/neutron/v4/testutil/dex/keeper"
	math_utils "github.com/neutron-org/neutron/v4/utils/math"
	"github.com/neutron-org/neutron/v4/x/dex/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.DexKeeper(t)
	params := types.DefaultParams()

	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	require.EqualValues(t, params, k.GetParams(ctx))
}

func TestValidateParams(t *testing.T) {
	goodFees := []uint64{1, 2, 3, 4, 5, 200}
	require.NoError(t, types.Params{FeeTiers: goodFees}.Validate())

	badFees := []uint64{1, 2, 3, 3}
	require.Error(t, types.Params{FeeTiers: badFees}.Validate())
}

func (s *DexTestSuite) TestPauseDex() {
	s.fundAliceBalances(100, 100)
	trancheKey := s.aliceLimitSells("TokenA", 0, 10, types.LimitOrderType_GOOD_TIL_CANCELLED)

	// WHEN params.paused is set to true
	params := types.DefaultParams()
	params.Paused = true
	_, err := s.msgServer.UpdateParams(s.Ctx, &types.MsgUpdateParams{Params: params, Authority: s.App.DexKeeper.GetAuthority()})

	s.NoError(err)

	// THEN all messages fail
	s.assertAliceDepositFails(types.ErrDexPaused, NewDeposit(0, 10, 0, 1))
	s.aliceWithdrawFails(types.ErrDexPaused, NewWithdrawal(5, 0, 1))
	s.assertAliceLimitSellFails(types.ErrDexPaused, "TokenB", -2, 1, types.LimitOrderType_IMMEDIATE_OR_CANCEL)
	s.aliceWithdrawLimitSellFails(types.ErrDexPaused, trancheKey)
	s.aliceCancelsLimitSellFails(trancheKey, types.ErrDexPaused)
	s.aliceMultiHopSwapFails(types.ErrDexPaused, [][]string{{"TokenA", "TokenB"}}, 5, math_utils.MustNewPrecDecFromStr("0.01"), false)

	// WHEN params.paused is set to false
	params.Paused = false
	_, err = s.msgServer.UpdateParams(s.Ctx, &types.MsgUpdateParams{Params: params, Authority: s.App.DexKeeper.GetAuthority()})
	s.NoError(err)

	// THEN all messages succeed
	s.aliceDeposits(NewDeposit(0, 10, 0, 1))
	s.aliceWithdraws(NewWithdrawal(5, 0, 1))
	s.aliceLimitSells("TokenB", -2, 1, types.LimitOrderType_IMMEDIATE_OR_CANCEL)
	s.aliceWithdrawsLimitSell(trancheKey)
	s.aliceCancelsLimitSell(trancheKey)
	s.aliceMultiHopSwaps([][]string{{"TokenA", "TokenB"}}, 5, math_utils.MustNewPrecDecFromStr("0.01"), false)
}
