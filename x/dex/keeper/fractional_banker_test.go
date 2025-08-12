package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v8/utils/math"
	"github.com/neutron-org/neutron/v8/x/dex/types"
)

const (
	TokenA = "TokenA"
	TokenB = "TokenB"
	TokenC = "TokenC"
)

func (s *DexTestSuite) TestFractionalBankerSendFractionalCoinsFromDexToAccount() {
	err := s.App.BankKeeper.MintCoins(s.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(TokenA, sdkmath.NewInt(100))))
	s.NoError(err)

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

	// send 0.3 => alice pays 1; dex gets 1; alice owed 0.71 (0.01 + 0.7)
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

func (s *DexTestSuite) TestFractionalBankerSendFractionalCoinsFromDexToAccountMultipleDenoms() {
	err := s.App.BankKeeper.MintCoins(s.Ctx, types.ModuleName,
		sdk.NewCoins(
			sdk.NewCoin(TokenA, sdkmath.NewInt(100)),
			sdk.NewCoin(TokenB, sdkmath.NewInt(100)),
			sdk.NewCoin(TokenC, sdkmath.NewInt(100)),
		))
	s.NoError(err)

	s.SendFractionalAmountToAccount("10.1", s.alice, TokenA)
	s.assertDexBalancesInt(sdkmath.NewInt(90), sdkmath.NewInt(100))
	s.assertAliceBalancesInt(sdkmath.NewInt(10), sdkmath.NewInt(0))
	s.assertFractionalBalance(s.alice, "0.1", "0")

	s.SendFractionalAmountToAccount("5.3", s.alice, TokenB)
	s.assertDexBalancesInt(sdkmath.NewInt(90), sdkmath.NewInt(95))
	s.assertAliceBalancesInt(sdkmath.NewInt(10), sdkmath.NewInt(5))
	s.assertFractionalBalance(s.alice, "0.1", "0.3")

	s.SendFractionalAmountToAccount("0.7", s.alice, TokenC)
	s.assertAccountBalanceWithDenomInt(s.alice, TokenC, sdkmath.NewInt(0))
	s.assertFractionalBalance(s.alice, "0.1", "0.3")
	s.assertFractionalBalanceCustomDenom(s.alice, TokenC, "0.7")

	s.SendFractionalAmountsToAccount(
		[]types.PrecDecCoin{
			types.NewPrecDecCoin(TokenA, math_utils.MustNewPrecDecFromStr("1.1")),
			types.NewPrecDecCoin(TokenB, math_utils.MustNewPrecDecFromStr("1.1")),
			types.NewPrecDecCoin(TokenC, math_utils.MustNewPrecDecFromStr("1.1")),
		},
		s.alice,
	)

	s.assertDexBalancesInt(sdkmath.NewInt(89), sdkmath.NewInt(94))
	s.assertDexBalanceWithDenomInt(TokenC, sdkmath.NewInt(99))
	s.assertAliceBalancesInt(sdkmath.NewInt(11), sdkmath.NewInt(6))
	s.assertAccountBalanceWithDenomInt(s.alice, TokenC, sdkmath.NewInt(1))
	s.assertFractionalBalance(s.alice, "0.2", "0.4")
	s.assertFractionalBalanceCustomDenom(s.alice, TokenC, "0.8")
}

func (s *DexTestSuite) TestFractionalBankerSendFractionalCoinsFromAccountToDexMultipleDenoms() {
	s.fundAccountBalancesInt(s.alice, sdkmath.NewInt(100), sdkmath.NewInt(100))
	s.fundAccountBalancesWithDenom(s.alice, sdk.NewCoins(sdk.NewCoin(TokenC, sdkmath.NewInt(100))))

	s.SendFractionalAmountFromAccount("10.1", s.alice, TokenA)
	s.assertDexBalancesInt(sdkmath.NewInt(11), sdkmath.NewInt(0))
	s.assertAliceBalancesInt(sdkmath.NewInt(89), sdkmath.NewInt(100))
	s.assertFractionalBalance(s.alice, "0.9", "0")

	s.SendFractionalAmountFromAccount("5.3", s.alice, TokenB)
	s.assertDexBalancesInt(sdkmath.NewInt(11), sdkmath.NewInt(6))
	s.assertAliceBalancesInt(sdkmath.NewInt(89), sdkmath.NewInt(94))
	s.assertFractionalBalance(s.alice, "0.9", "0.7")

	s.SendFractionalAmountFromAccount("0.7", s.alice, TokenC)
	s.assertAccountBalanceWithDenomInt(s.alice, TokenC, sdkmath.NewInt(99))
	s.assertFractionalBalance(s.alice, "0.9", "0.7")
	s.assertFractionalBalanceCustomDenom(s.alice, TokenC, "0.3")

	s.SendFractionalAmountsFromAccount(
		[]types.PrecDecCoin{
			types.NewPrecDecCoin(TokenA, math_utils.MustNewPrecDecFromStr("1.1")),
			types.NewPrecDecCoin(TokenB, math_utils.MustNewPrecDecFromStr("1.8")),
			types.NewPrecDecCoin(TokenC, math_utils.MustNewPrecDecFromStr("1.1")),
		},
		s.alice,
	)
	s.assertDexBalancesInt(sdkmath.NewInt(12), sdkmath.NewInt(8))
	s.assertDexBalanceWithDenomInt(TokenC, sdkmath.NewInt(2))
	s.assertAliceBalancesInt(sdkmath.NewInt(88), sdkmath.NewInt(92))
	s.assertAccountBalanceWithDenomInt(s.alice, TokenC, sdkmath.NewInt(98))
	s.assertFractionalBalance(s.alice, "0.8", "0.9")
	s.assertFractionalBalanceCustomDenom(s.alice, TokenC, "0.2")
}

func (s *DexTestSuite) SendFractionalAmountToAccount(amount string, account sdk.AccAddress, denom ...string) {
	sendDenom := TokenA
	if len(denom) > 0 {
		sendDenom = denom[0]
	}

	s.SendFractionalAmountsToAccount(
		[]types.PrecDecCoin{types.NewPrecDecCoin(sendDenom, math_utils.MustNewPrecDecFromStr(amount))},
		account,
	)
}

func (s *DexTestSuite) SendFractionalAmountsFromAccount(amounts types.PrecDecCoins, account sdk.AccAddress) {
	err := s.App.DexKeeper.SendFractionalCoinsFromAccountToDex(
		s.Ctx,
		account,
		amounts,
	)
	s.NoError(err)
}

func (s *DexTestSuite) SendFractionalAmountFromAccount(amount string, account sdk.AccAddress, denom ...string) {
	sendDenom := TokenA
	if len(denom) > 0 {
		sendDenom = denom[0]
	}

	s.SendFractionalAmountsFromAccount(
		[]types.PrecDecCoin{types.NewPrecDecCoin(sendDenom, math_utils.MustNewPrecDecFromStr(amount))},
		account,
	)
}

func (s *DexTestSuite) SendFractionalAmountsToAccount(amounts types.PrecDecCoins, account sdk.AccAddress) {
	err := s.App.DexKeeper.SendFractionalCoinsFromDexToAccount(
		s.Ctx,
		account,
		amounts,
	)
	s.NoError(err)
}

func (s *DexTestSuite) assertFractionalBalance(account sdk.AccAddress, expectedAmountA, expectedAmountB string) {
	balance, err := s.App.DexKeeper.GetFractionalBalances(s.Ctx, account, TokenA, TokenB)
	s.NoError(err)
	tokenABalance := balance.AmountOf(TokenA)
	tokenBBalance := balance.AmountOf(TokenB)

	s.Require().Equal(math_utils.MustNewPrecDecFromStr(expectedAmountA), tokenABalance, "Expected balance A %v != %v", expectedAmountA, tokenABalance.String())
	s.Require().Equal(math_utils.MustNewPrecDecFromStr(expectedAmountB), tokenBBalance, "Expected balance B %v != %v", expectedAmountB, tokenBBalance.String())
}

func (s *DexTestSuite) assertFractionalBalanceCustomDenom(
	account sdk.AccAddress,
	denom, expectedAmount string, //nolint:unparam
) {
	balance, err := s.App.DexKeeper.GetFractionalBalances(s.Ctx, account, denom)
	s.NoError(err)
	tokenABalance := balance.AmountOf(denom)

	s.Require().Equal(math_utils.MustNewPrecDecFromStr(expectedAmount), tokenABalance, "Expected balance A %v != %v", expectedAmount, tokenABalance.String())
}
