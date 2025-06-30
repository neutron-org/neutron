package keeper

import (
	"errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	math_utils "github.com/neutron-org/neutron/v7/utils/math"
	"github.com/neutron-org/neutron/v7/x/dex/types"
)

type FractionalBanker struct {
	AmountsOwed collections.Map[string, PrecDecCoins]
	BankKeeper  types.BankKeeper
}

func NewFractionalBanker(storeService store.KVStoreService, bankKeeper types.BankKeeper, cdc codec.Codec) *FractionalBanker {
	sb := collections.NewSchemaBuilder(storeService)

	return &FractionalBanker{
		BankKeeper: bankKeeper,
		AmountsOwed: collections.NewMap(
			sb,
			collections.NewPrefix(types.AmountsOwedKey),
			"AmountsOwed",
			collections.StringKey,
			codec.CollValue[types.PrecDecCoins](cdc),
		),
	}
}

func (k *FractionalBanker) SendFractionalToken(ctx sdk.Context, address string, tokens []types.PrecDecCoin) error {
	amountsOwed, err := k.AmountsOwed.Get(ctx, address)
	if errors.Is(err, collections.ErrNotFound) {
		amountsOwed = types.PrecDecCoins{}
	} else if err != nil {
		return err
	}

	amountsOwed = amountsOwed.Add(tokens...)

}

func GetWholeTokenAmounts(tokens []types.PrecDecCoin) (wholeTokens types.Coins, fractionalTokens types.PrecDecCoins) {
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
