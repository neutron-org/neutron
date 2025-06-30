package types

import (
	fmt "fmt"
	"sort"

	math_utils "github.com/neutron-org/neutron/v7/utils/math"
)

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
