package utils

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v3/x/dex/types"

	"github.com/cosmos/cosmos-sdk/types/query"
)

func FilteredPaginateAccountBalances(
	ctx sdk.Context,
	bankKeeper types.BankKeeper,
	address sdk.AccAddress,
	pageRequest *query.PageRequest,
	onResult func(coin sdk.Coin, accumulate bool) bool,
) (*query.PageResponse, error) {
	// if the PageRequest is nil, use default PageRequest
	if pageRequest == nil {
		pageRequest = &query.PageRequest{}
	}

	offset := pageRequest.Offset
	key := pageRequest.Key
	limit := pageRequest.Limit
	countTotal := pageRequest.CountTotal

	if pageRequest.Reverse {
		return nil, fmt.Errorf("invalid request, reverse pagination is not enabled")
	}
	if offset > 0 && key != nil {
		return nil, fmt.Errorf("invalid request, either offset or key is expected, got both")
	}

	if limit == 0 {
		limit = query.DefaultLimit

		// count total results when the limit is zero/not supplied
		countTotal = true
	}

	if len(key) != 0 {
		// paginate with key
		var (
			numHits uint64
			nextKey []byte
		)
		startAccum := false

		bankKeeper.IterateAccountBalances(ctx, address, func(coin sdk.Coin) bool {
			if coin.Denom == string(key) {
				startAccum = true
			}
			if numHits == limit {
				nextKey = []byte(coin.Denom)
				return true
			}
			if startAccum {
				hit := onResult(coin, true)
				if hit {
					numHits++
				}
			}

			return false
		})

		return &query.PageResponse{
			NextKey: nextKey,
		}, nil
	}
	// else  default pagination (with offset)
	end := offset + limit
	var (
		numHits uint64
		nextKey []byte
	)

	bankKeeper.IterateAccountBalances(ctx, address, func(coin sdk.Coin) bool {
		accumulate := numHits >= offset && numHits < end
		hit := onResult(coin, accumulate)

		if hit {
			numHits++
		}

		if numHits == end+1 {
			if nextKey == nil {
				nextKey = []byte(coin.Denom)
			}

			if !countTotal {
				return true
			}
		}

		return false
	})

	res := &query.PageResponse{NextKey: nextKey}
	if countTotal {
		res.Total = numHits
	}

	return res, nil
}

// SanitizeCoins takes an unsorted list of coins and sorts them, removes coins with amount zero and combines duplicate coins
func SanitizeCoins(coins []sdk.Coin) sdk.Coins {
	sort.SliceStable(coins, func(i, j int) bool {
		return coins[i].Denom < coins[j].Denom
	})
	cleanCoins := sdk.Coins{}
	lastDenom := ""
	for _, coin := range coins {
		if coin.IsZero() {
			continue
		}
		if lastDenom != coin.Denom {
			cleanCoins = append(cleanCoins, coin)
		} else {
			cleanCoins[len(cleanCoins)-1].Add(coin)
		}
		lastDenom = coin.Denom
	}
	return cleanCoins
}
