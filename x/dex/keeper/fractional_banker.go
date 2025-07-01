package keeper

import (
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
		return []types.PrecDecCoin{}
	}

	var balance types.FractionalBalance
	k.cdc.MustUnmarshal(b, &balance)

	return balance.Balance
}

func (k *FractionalBanker) SetFractionalBalance(ctx sdk.Context, address sdk.AccAddress, coins types.PrecDecCoins) {

	balance := types.FractionalBalance{
		Balance: coins,
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FractionalBalanceKeyPrefix))
	b := k.cdc.MustMarshal(&balance)
	store.Set(types.FractionalBalanceKey(address), b)
}

func (k *FractionalBanker) SendFractionalToken(ctx sdk.Context, address sdk.AccAddress, tokens []types.PrecDecCoin) error {
	var balance types.PrecDecCoins = k.GetFractionalBalance(ctx, address)

	newBalance := balance.Add(tokens...)

	wholeTokens, fractionalTokens := GetWholeTokenAmounts(newBalance)

	if !wholeTokens.Empty() {
		err := k.BankKeeper.SendCoinsFromAccountToModule(ctx, address, types.ModuleName, wholeTokens)
		if err != nil {
			return err
		}
	}

	k.SetFractionalBalance(ctx, address, fractionalTokens)

	return nil
}

func GetWholeTokenAmounts(tokens types.PrecDecCoins) (wholeTokens sdk.Coins, fractionalTokens types.PrecDecCoins) {
	wholeTokens = sdk.Coins{}
	fractionalTokens = types.PrecDecCoins{}

	for _, token := range tokens {
		wholeAmount := token.Amount.TruncateInt()
		fractionalRemainder := token.Amount.Sub(math_utils.NewPrecDecFromInt(wholeAmount))
		if !wholeAmount.IsZero() {
			wholeTokens = append(wholeTokens, sdk.Coin{Denom: token.Denom, Amount: wholeAmount})
		}
		if !fractionalRemainder.IsZero() {
			fractionalTokens = append(fractionalTokens, types.PrecDecCoin{Denom: token.Denom, Amount: fractionalRemainder})
		}
	}

	return wholeTokens, fractionalTokens
}
