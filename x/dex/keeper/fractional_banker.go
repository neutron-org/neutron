package keeper

import (
	math "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v7/utils/math"
	"github.com/neutron-org/neutron/v7/x/dex/types"
)

type FractionalBanker struct {
	BankKeeper types.BankKeeper
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
}

func NewFractionalBanker(storeKey storetypes.StoreKey, bankKeeper types.BankKeeper, cdc codec.BinaryCodec) *FractionalBanker {
	return &FractionalBanker{
		BankKeeper: bankKeeper,
		storeKey:   storeKey,
		cdc:        cdc,
	}
}

func (k *FractionalBanker) GetFractionalBalance(ctx sdk.Context, address sdk.AccAddress) types.PrecDecCoins {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FractionalBalanceKeyPrefix))
	b := store.Get(types.FractionalBalanceKey(address))

	if b == nil {
		return types.NewPrecDecCoins()
	}

	var balance types.FractionalBalance
	k.cdc.MustUnmarshal(b, &balance)

	return balance.Balance
}

func (k *FractionalBanker) GetAllFractionalBalances(ctx sdk.Context) types.PrecDecCoins {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FractionalBalanceKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	balances := types.PrecDecCoins{}

	for ; iterator.Valid(); iterator.Next() {
		var balance types.FractionalBalance
		k.cdc.MustUnmarshal(iterator.Value(), &balance)
		balances = balances.Add(balance.Balance...)
	}

	return balances
}

func (k *FractionalBanker) SetFractionalBalance(ctx sdk.Context, address sdk.AccAddress, coins types.PrecDecCoins) {
	balance := types.FractionalBalance{
		Balance: coins,
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FractionalBalanceKeyPrefix))
	b := k.cdc.MustMarshal(&balance)
	store.Set(types.FractionalBalanceKey(address), b)
}

func (k *FractionalBanker) SendFractionalCoinsFromDexToAccount(ctx sdk.Context, address sdk.AccAddress, tokens []types.PrecDecCoin) error {
	balance := k.GetFractionalBalance(ctx, address)

	newBalance := balance.Add(tokens...)

	wholeTokens, fractionalTokens := RoundDownToWholeTokenAmounts(newBalance)

	if !wholeTokens.Empty() {
		err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, wholeTokens)
		if err != nil {
			return err
		}
	}

	k.SetFractionalBalance(ctx, address, fractionalTokens)

	return nil
}

func (k *FractionalBanker) SendFractionalCoinsFromAccountToDex(ctx sdk.Context, address sdk.AccAddress, tokens []types.PrecDecCoin) error {
	balance := k.GetFractionalBalance(ctx, address)

	coinsToSend, newDebts := CalcUserSendMinusDebts(tokens, balance)

	if !coinsToSend.Empty() {
		err := k.BankKeeper.SendCoinsFromAccountToModule(ctx, address, types.ModuleName, coinsToSend)
		if err != nil {
			return err
		}
	}

	k.SetFractionalBalance(ctx, address, newDebts)

	return nil
}

func RoundDownToWholeTokenAmounts(tokens types.PrecDecCoins) (wholeTokens sdk.Coins, fractionalTokens types.PrecDecCoins) {
	wholeTokens = sdk.Coins{}
	fractionalTokens = types.PrecDecCoins{}

	for _, token := range tokens {
		wholeAmount := token.Amount.TruncateInt()
		fractionalRemainder := token.Amount.Sub(math_utils.NewPrecDecFromInt(wholeAmount))
		if !wholeAmount.IsZero() {
			wholeTokens = wholeTokens.Add(sdk.Coin{Denom: token.Denom, Amount: wholeAmount})
		}
		if !fractionalRemainder.IsZero() {
			fractionalTokens = fractionalTokens.Add(types.NewPrecDecCoin(token.Denom, fractionalRemainder))
		}
	}

	return wholeTokens, fractionalTokens
}

func CalcUserSendMinusDebts(amountToSend types.PrecDecCoins, debts types.PrecDecCoins) (sdk.Coins, types.PrecDecCoins) {
	coinsToSend := sdk.NewCoins()
	for _, coinToPay := range amountToSend {
		var userPays math.Int
		var remainingDebt math_utils.PrecDec
		debtAmount := debts.AmountOf(coinToPay.Denom)
		if coinToPay.Amount.LTE(debtAmount) {
			// Use outstanding debt to cover the amount the user is paying
			userPays = math.ZeroInt()
			// reduce debt by the amount applied to the balance
			remainingDebt = debtAmount.Sub(coinToPay.Amount)
		} else {
			// Subtract debt from the amount the user is paying
			userPaysRaw := coinToPay.Amount.Sub(debtAmount)
			// round up to the nearest whole number
			userPays = userPaysRaw.Ceil().TruncateInt()
			// remaining debt is the difference between the rounded up amount and the original amount
			remainingDebt = userPaysRaw.Ceil().Sub(userPaysRaw)
		}
		coinsToSend = coinsToSend.Add(sdk.NewCoin(coinToPay.Denom, userPays))
		debts = debts.SetAmountOf(types.NewPrecDecCoin(coinToPay.Denom, remainingDebt))
	}
	return coinsToSend, debts
}

func RoundUpToWholeTokenAmounts(tokens types.PrecDecCoins) (wholeTokens sdk.Coins, fractionalTokens types.PrecDecCoins) {
	wholeTokens = sdk.Coins{}
	fractionalTokens = types.PrecDecCoins{}

	for _, token := range tokens {
		wholeAmount := token.Amount.Ceil().TruncateInt()
		fractionalRemainder := math_utils.NewPrecDecFromInt(wholeAmount).Sub(token.Amount)
		if !wholeAmount.IsZero() {
			wholeTokens = append(wholeTokens, sdk.Coin{Denom: token.Denom, Amount: wholeAmount})
		}
		if !fractionalRemainder.IsZero() {
			fractionalTokens = append(fractionalTokens, types.PrecDecCoin{Denom: token.Denom, Amount: fractionalRemainder})
		}
	}

	return wholeTokens, fractionalTokens
}
