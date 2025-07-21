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

type (
	DebtMap          map[string]math_utils.PrecDec
	FractionalBanker struct {
		BankKeeper types.BankKeeper
		storeKey   storetypes.StoreKey
		cdc        codec.BinaryCodec
	}
)

func NewFractionalBanker(storeKey storetypes.StoreKey, bankKeeper types.BankKeeper, cdc codec.BinaryCodec) *FractionalBanker {
	return &FractionalBanker{
		BankKeeper: bankKeeper,
		storeKey:   storeKey,
		cdc:        cdc,
	}
}

func (k *FractionalBanker) GetFractionalBalances(ctx sdk.Context, address sdk.AccAddress, denoms ...string) types.PrecDecCoins {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FractionalBalanceKeyPrefix))
	balance := types.PrecDecCoins{}
	for _, denom := range denoms {
		b := store.Get(types.FractionalBalanceKey(address, denom))

		if b == nil {
			balance = balance.Add(types.NewPrecDecCoin(denom, math_utils.ZeroPrecDec()))
		} else {
			amount := math_utils.MustNewPrecDecFromStr(string(b))
			balance = balance.Add(types.NewPrecDecCoin(denom, amount))
		}
	}

	return balance
}

func (k *FractionalBanker) GetAllFractionalBalances(ctx sdk.Context) types.PrecDecCoins {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FractionalBalanceKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	balances := types.PrecDecCoins{}

	for ; iterator.Valid(); iterator.Next() {
		denom := string(iterator.Key()[len(types.FractionalBalanceKeyPrefix):])
		amount := math_utils.MustNewPrecDecFromStr(string(iterator.Value()))
		balances = balances.Add(types.NewPrecDecCoin(denom, amount))
	}

	return balances
}

func (k *FractionalBanker) SetFractionalBalanceFromMap(ctx sdk.Context, address sdk.AccAddress, debtMap DebtMap) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FractionalBalanceKeyPrefix))
	for denom, amount := range debtMap {
		if amount.IsPositive() {
			store.Set(types.FractionalBalanceKey(address, denom), []byte(amount.String()))
		} else {
			store.Delete(types.FractionalBalanceKey(address, denom))
		}
	}
}

func (k *FractionalBanker) SetFractionalBalance(ctx sdk.Context, address sdk.AccAddress, coins types.PrecDecCoins) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FractionalBalanceKeyPrefix))
	for _, coin := range coins {
		b := []byte(coin.Amount.String())
		if coin.Amount.IsZero() {
			store.Delete(types.FractionalBalanceKey(address, coin.Denom))
			continue
		} // else
		store.Set(types.FractionalBalanceKey(address, coin.Denom), b)
	}
}

func (k *FractionalBanker) SendFractionalCoinsFromDexToAccount(ctx sdk.Context, address sdk.AccAddress, tokens []types.PrecDecCoin) error {
	relevantDenoms := make([]string, 0)
	for _, coin := range tokens {
		relevantDenoms = append(relevantDenoms, coin.Denom)
	}
	balance := k.GetFractionalBalances(ctx, address, relevantDenoms...)

	newBalance := balance.Add(tokens...)

	wholeTokens, fractionalDebts := RoundDownToWholeTokenAmounts(newBalance)

	if !wholeTokens.Empty() {
		err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, wholeTokens)
		if err != nil {
			return err
		}
	}

	k.SetFractionalBalanceFromMap(ctx, address, fractionalDebts)

	return nil
}

func (k *FractionalBanker) SendFractionalCoinsFromAccountToDex(ctx sdk.Context, address sdk.AccAddress, tokens []types.PrecDecCoin) error {
	var relevantDenoms []string
	for _, coin := range tokens {
		relevantDenoms = append(relevantDenoms, coin.Denom)
	}
	balances := k.GetFractionalBalances(ctx, address, relevantDenoms...)

	coinsToSend, debtMap := CalcUserSendMinusDebts(tokens, balances)

	if !coinsToSend.Empty() {
		err := k.BankKeeper.SendCoinsFromAccountToModule(ctx, address, types.ModuleName, coinsToSend)
		if err != nil {
			return err
		}
	}

	k.SetFractionalBalanceFromMap(ctx, address, debtMap)

	return nil
}

func RoundDownToWholeTokenAmounts(tokens types.PrecDecCoins) (wholeTokens sdk.Coins, fractionalDebts DebtMap) {
	wholeTokens = sdk.Coins{}
	fractionalDebts = make(DebtMap)

	for _, token := range tokens {
		wholeAmount := token.Amount.TruncateInt()
		fractionalRemainder := token.Amount.Sub(math_utils.NewPrecDecFromInt(wholeAmount))
		if !wholeAmount.IsZero() {
			wholeTokens = wholeTokens.Add(sdk.Coin{Denom: token.Denom, Amount: wholeAmount})
			fractionalDebts[token.Denom] = math_utils.ZeroPrecDec()
		}
		if !fractionalRemainder.IsZero() {
			fractionalDebts[token.Denom] = fractionalRemainder
		}
	}

	return wholeTokens, fractionalDebts
}

func CalcUserSendMinusDebts(amountToSend, debts types.PrecDecCoins) (sdk.Coins, DebtMap) {
	coinsToSend := sdk.NewCoins()
	debtMap := make(DebtMap)
	for _, coinToPay := range amountToSend {
		var userPays math.Int
		debtAmount := debts.AmountOf(coinToPay.Denom)
		if coinToPay.Amount.LTE(debtAmount) {
			// Use outstanding debt to cover the amount the user is paying
			userPays = math.ZeroInt()
			// reduce debt by the amount applied to the balance
			debtMap[coinToPay.Denom] = debtAmount.Sub(coinToPay.Amount)
		} else {
			// Subtract debt from the amount the user is paying
			userPaysRaw := coinToPay.Amount.Sub(debtAmount)
			// round up to the nearest whole number
			userPays = userPaysRaw.Ceil().TruncateInt()
			// remaining debt is the difference between the rounded up amount and the original amount
			debtMap[coinToPay.Denom] = userPaysRaw.Ceil().Sub(userPaysRaw)
		}
		coinsToSend = coinsToSend.Add(sdk.NewCoin(coinToPay.Denom, userPays))

	}
	return coinsToSend, debtMap
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
