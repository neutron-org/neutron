package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	math_utils "github.com/neutron-org/neutron/v7/utils/math"
	"github.com/neutron-org/neutron/v7/x/dex/types"
)

func (s *DexTestSuite) TestFractionalBankerSendFractionalCoinsFromDexToAccount() {
	s.App.BankKeeper.MintCoins(s.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin("TokenA", sdkmath.NewInt(1000000000000))))

	// send 10.1 => alice gets 10; owed 0.1
	s.SendFractionalAmountToAccount("10.1", s.alice)
	s.assertAliceBalancesInt(sdkmath.NewInt(10), sdkmath.NewInt(0))
	s.assertFractionalBalance(s.alice, "0.1", "0")

	// send 5.3 => alice gets 5; owed 0.4
	s.SendFractionalAmountToAccount("5.3", s.alice)
	s.assertAliceBalancesInt(sdkmath.NewInt(15), sdkmath.NewInt(0))
	s.assertFractionalBalance(s.alice, "0.4", "0")

	// send 0.7 => alice gets 1; owed 0.1
	s.SendFractionalAmountToAccount("0.7", s.alice)
	s.assertAliceBalancesInt(sdkmath.NewInt(16), sdkmath.NewInt(0))
	s.assertFractionalBalance(s.alice, "0.1", "0")

	// send 0.9 => alice gets 1; owed 0
	s.SendFractionalAmountToAccount("0.9", s.alice)
	s.assertAliceBalancesInt(sdkmath.NewInt(17), sdkmath.NewInt(0))
	s.assertFractionalBalance(s.alice, "0", "0")

}

func (s *DexTestSuite) TestFractionalBankerSendFractionalCoinsFromAccountToDex() {
	s.fundAccountBalancesInt(s.alice, sdkmath.NewInt(100), sdkmath.NewInt(0))

	// send 5 => alice pays 5; dex gets 5; alice owed 0
	s.SendFractionalAmountFromAccount("5", s.alice)
	s.assertDexBalancesInt(sdkmath.NewInt(5), sdkmath.NewInt(0))
	s.assertFractionalBalance(s.alice, "0", "0")
	s.assertAliceBalancesInt(sdkmath.NewInt(95), sdkmath.NewInt(0))

	// send 5.99 => alice pays 6; dex gets 6; alice owed 0.01
	s.SendFractionalAmountFromAccount("5.99", s.alice)
	s.assertDexBalancesInt(sdkmath.NewInt(11), sdkmath.NewInt(0))
	s.assertFractionalBalance(s.alice, "0.01", "0")
	s.assertAliceBalancesInt(sdkmath.NewInt(89), sdkmath.NewInt(0))

	// send 0.3 => alice pays 1; dex gets 1; alice owed 0.31
	s.SendFractionalAmountFromAccount("0.3", s.alice)
	s.assertDexBalancesInt(sdkmath.NewInt(12), sdkmath.NewInt(0))
	s.assertFractionalBalance(s.alice, "0.71", "0")
	s.assertAliceBalancesInt(sdkmath.NewInt(88), sdkmath.NewInt(0))

	// send 0.1 => alice pays 0; dex gets 0; alice owed 0.61
	s.SendFractionalAmountFromAccount("0.1", s.alice)
	s.assertDexBalancesInt(sdkmath.NewInt(12), sdkmath.NewInt(0))
	s.assertFractionalBalance(s.alice, "0.61", "0")
	s.assertAliceBalancesInt(sdkmath.NewInt(88), sdkmath.NewInt(0))

	// send 10.2 => alice pays 10; dex gets 10; alice owed 0.41
	s.SendFractionalAmountFromAccount("10.2", s.alice)
	s.assertDexBalancesInt(sdkmath.NewInt(22), sdkmath.NewInt(0))
	s.assertFractionalBalance(s.alice, "0.41", "0")
	s.assertAliceBalancesInt(sdkmath.NewInt(78), sdkmath.NewInt(0))

	// send 0.41 => alice pays 0; dex gets 0; alice owed 0
	s.SendFractionalAmountFromAccount("0.41", s.alice)
	s.assertDexBalancesInt(sdkmath.NewInt(22), sdkmath.NewInt(0))
	s.assertFractionalBalance(s.alice, "0", "0")
	s.assertAliceBalancesInt(sdkmath.NewInt(78), sdkmath.NewInt(0))

}

func (s *DexTestSuite) SendFractionalAmountToAccount(amount string, account sdk.AccAddress) {
	s.App.DexKeeper.SendFractionalCoinsFromDexToAccount(
		s.Ctx,
		account,
		[]types.PrecDecCoin{types.NewPrecDecCoin("TokenA", math_utils.MustNewPrecDecFromStr(amount))},
	)

}

func (s *DexTestSuite) SendFractionalAmountFromAccount(amount string, account sdk.AccAddress) {
	s.App.DexKeeper.SendFractionalCoinsFromAccountToDex(
		s.Ctx,
		account,
		[]types.PrecDecCoin{types.NewPrecDecCoin("TokenA", math_utils.MustNewPrecDecFromStr(amount))},
	)

}

func (s *DexTestSuite) assertFractionalBalance(account sdk.AccAddress, expectedAmountA, expectedAmountB string) {
	balance := s.App.DexKeeper.GetFractionalBalances(s.Ctx, account, "TokenA", "TokenB")
	tokenABalance := balance.AmountOf("TokenA")
	tokenBBalance := balance.AmountOf("TokenB")

	s.Require().Equal(math_utils.MustNewPrecDecFromStr(expectedAmountA), tokenABalance, "Expected balance A %v != %v", expectedAmountA, tokenABalance.String())
	s.Require().Equal(math_utils.MustNewPrecDecFromStr(expectedAmountB), tokenBBalance, "Expected balance B %v != %v", expectedAmountB, tokenBBalance.String())
}
