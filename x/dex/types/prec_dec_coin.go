package types

import (
	"errors"
	fmt "fmt"
	"sort"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	math_utils "github.com/neutron-org/neutron/v7/utils/math"
)

func NewPrecDecCoin(denom string, amount math_utils.PrecDec) PrecDecCoin {
	coin := PrecDecCoin{
		Denom:  denom,
		Amount: amount,
	}

	if err := coin.Validate(); err != nil {
		panic(err)
	}

	return coin
}

func NewPrecDecCoinFromInt(denom string, amount math.Int) PrecDecCoin {
	coin := PrecDecCoin{
		Denom:  denom,
		Amount: math_utils.NewPrecDecFromInt(amount),
	}

	if err := coin.Validate(); err != nil {
		panic(err)
	}

	return coin
}

func NewPrecDecCoinFromCoin(coin sdk.Coin) PrecDecCoin {
	precDecCoin := PrecDecCoin{
		Denom:  coin.Denom,
		Amount: math_utils.NewPrecDecFromInt(coin.Amount),
	}

	if err := coin.Validate(); err != nil {
		panic(err)
	}

	return precDecCoin
}

func (coin PrecDecCoin) String() string {
	return fmt.Sprintf("%s%s", coin.Amount.String(), coin.Denom)
}

func (coin PrecDecCoin) TruncateToCoin() sdk.Coin {
	return sdk.Coin{
		Denom:  coin.Denom,
		Amount: coin.Amount.TruncateInt(), // TODO: check if this is correct
	}
}

func (coin PrecDecCoin) CeilToCoin() sdk.Coin {
	return sdk.Coin{
		Denom:  coin.Denom,
		Amount: coin.Amount.Ceil().TruncateInt(), // TODO: check if this is correct
	}
}

// Validate returns an error if the Coin has a negative amount or if
// the denom is invalid.
func (coin PrecDecCoin) Validate() error {
	if err := sdk.ValidateDenom(coin.Denom); err != nil {
		return err
	}

	if coin.Amount.IsNil() {
		return errors.New("amount is nil")
	}

	if coin.Amount.IsNegative() {
		return fmt.Errorf("negative coin amount: %v", coin.Amount)
	}

	return nil
}

func (coin PrecDecCoin) Add(coinB PrecDecCoin) PrecDecCoin {
	if coin.Denom != coinB.Denom {
		panic(fmt.Sprintf("invalid coin denominations; %s, %s", coin.Denom, coinB.Denom))
	}

	return PrecDecCoin{coin.Denom, coin.Amount.Add(coinB.Amount)}
}

type PrecDecCoins []PrecDecCoin

func (coins PrecDecCoins) Add(coinsB ...PrecDecCoin) PrecDecCoins {
	return coins.safeAdd(coinsB)
}

func (coins PrecDecCoins) isSorted() bool {
	for i := 1; i < len(coins); i++ {
		if coins[i-1].Denom > coins[i].Denom {
			return false
		}
	}
	return true
}

// IsZero returns if this represents no money
func (coin PrecDecCoin) IsZero() bool {
	return coin.Amount.IsZero()
}

func (coin PrecDecCoin) IsPositive() bool {
	return coin.Amount.IsPositive()
}

func (coin PrecDecCoin) IsNegative() bool {
	return coin.Amount.IsNegative()
}

//-----------------------------------------------------------------------------
// Sort interface

// Len implements sort.Interface for Coins
func (coins PrecDecCoins) Len() int { return len(coins) }

// Less implements sort.Interface for Coins
func (coins PrecDecCoins) Less(i, j int) bool { return coins[i].Denom < coins[j].Denom }

// Swap implements sort.Interface for Coins
func (coins PrecDecCoins) Swap(i, j int) { coins[i], coins[j] = coins[j], coins[i] }

var _ sort.Interface = PrecDecCoins{}

func (coins PrecDecCoins) Sort() PrecDecCoins {
	// sort.Sort(coins) does a costly runtime copy as part of `runtime.convTSlice`
	// So we avoid this heap allocation if len(coins) <= 1. In the future, we should hopefully find
	// a strategy to always avoid this.
	if len(coins) > 1 {
		sort.Sort(coins)
	}
	return coins
}

//-----------------------------------------------------------------------------

func (coins PrecDecCoins) safeAdd(coinsB PrecDecCoins) (coalesced PrecDecCoins) {
	// probably the best way will be to make Coins and interface and hide the structure
	// definition (type alias)
	if !coins.isSorted() {
		panic("Coins (self) must be sorted")
	}
	if !coinsB.isSorted() {
		panic("Wrong argument: coins must be sorted")
	}

	uniqCoins := make(map[string]PrecDecCoins, len(coins)+len(coinsB))
	// Traverse all the coins for each of the coins and coinsB.
	for _, cL := range []PrecDecCoins{coins, coinsB} {
		for _, c := range cL {
			uniqCoins[c.Denom] = append(uniqCoins[c.Denom], c)
		}
	}

	for denom, cL := range uniqCoins { //#nosec
		comboCoin := PrecDecCoin{Denom: denom, Amount: math_utils.NewPrecDec(0)}
		for _, c := range cL {
			comboCoin = comboCoin.Add(c)
		}
		if !comboCoin.IsZero() {
			coalesced = append(coalesced, comboCoin)
		}
	}
	if coalesced == nil {
		return PrecDecCoins{}
	}
	return coalesced.Sort()
}

// Empty returns true if there are no coins and false otherwise.
func (coins PrecDecCoins) Empty() bool {
	return len(coins) == 0
}

func (coins PrecDecCoins) TruncateToCoins() sdk.Coins {
	truncatedCoins := make(sdk.Coins, len(coins))
	for i, coin := range coins {
		truncatedCoins[i] = coin.TruncateToCoin()
	}
	return truncatedCoins
}

func (coins PrecDecCoins) String() string {
	if len(coins) == 0 {
		return ""
	} else if len(coins) == 1 {
		return coins[0].String()
	}

	// Build the string with a string builder
	var out strings.Builder
	for _, coin := range coins[:len(coins)-1] {
		out.WriteString(coin.String())
		out.WriteByte(',')
	}
	out.WriteString(coins[len(coins)-1].String())
	return out.String()
}

func (coins PrecDecCoins) AmountOf(denom string) math_utils.PrecDec {
	if ok, c := coins.Find(denom); ok {
		return c.Amount
	}
	return math_utils.ZeroPrecDec()
}

// Find returns true and coin if the denom exists in coins. Otherwise it returns false
// and a zero coin. Uses binary search.
// CONTRACT: coins must be valid (sorted).
func (coins PrecDecCoins) Find(denom string) (bool, PrecDecCoin) {
	switch len(coins) {
	case 0:
		return false, PrecDecCoin{}

	case 1:
		coin := coins[0]
		if coin.Denom == denom {
			return true, coin
		}
		return false, PrecDecCoin{}

	default:
		midIdx := len(coins) / 2 // 2:1, 3:1, 4:2
		coin := coins[midIdx]
		switch {
		case denom < coin.Denom:
			return coins[:midIdx].Find(denom)
		case denom == coin.Denom:
			return true, coin
		default:
			return coins[midIdx+1:].Find(denom)
		}
	}
}
