package types

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/types/query"
	"golang.org/x/exp/maps"
)

/*
File contains an implementation of query responses custom json marshaller for wasmbinding
*/

type LimitOrderTrancheBinding struct {
	LimitOrderTranche
	ExpirationTime *uint64 `json:"expiration_time,omitempty"`
}

type LimitOrderTrancheUserBinding struct {
	LimitOrderTrancheUser
	OrderType string `json:"order_type"`
}

func (t LimitOrderTranche) ToBinding() LimitOrderTrancheBinding {
	lo := LimitOrderTrancheBinding{
		LimitOrderTranche: t,
	}
	if t.ExpirationTime != nil && t.ExpirationTime.Unix() >= 0 {
		ut := uint64(t.ExpirationTime.Unix()) //nolint:gosec
		lo.ExpirationTime = &ut
	}
	return lo
}

func (t *QueryGetPoolMetadataResponse) MarshalBinding() ([]byte, error) {
	type QueryGetPoolMetadataResponseBinding struct {
		PoolMetadata PoolMetadata `json:"pool_metadata"`
	}

	metadata := QueryGetPoolMetadataResponseBinding(*t)
	return json.Marshal(&metadata)
}

func (t *QueryAllTickLiquidityResponse) MarshalBinding() ([]byte, error) {
	type tickLiquidity struct {
		Liquidity struct {
			LimitOrderTranche *LimitOrderTrancheBinding `json:"limit_order_tranche,omitempty"`
			PoolReserves      *PoolReserves             `json:"pool_reserves,omitempty"`
		} `json:"liquidity,omitempty"`
	}
	type QueryAllTickLiquidityResponseBinding struct {
		TickLiquidity []tickLiquidity     `json:"tick_liquidity,omitempty"`
		Pagination    *query.PageResponse `json:"pagination,omitempty"`
	}
	q := QueryAllTickLiquidityResponseBinding{
		TickLiquidity: make([]tickLiquidity, 0, len(t.TickLiquidity)),
		Pagination:    t.Pagination,
	}
	for _, tl := range t.TickLiquidity {
		if lq, ok := tl.Liquidity.(*TickLiquidity_LimitOrderTranche); ok {
			lo := lq.LimitOrderTranche.ToBinding()
			tlNew := tickLiquidity{}
			tlNew.Liquidity.LimitOrderTranche = &lo
			q.TickLiquidity = append(q.TickLiquidity, tlNew)
		} else {
			tlNew := tickLiquidity{}
			tlNew.Liquidity.PoolReserves = tl.Liquidity.(*TickLiquidity_PoolReserves).PoolReserves
			q.TickLiquidity = append(q.TickLiquidity, tlNew)
		}
	}
	return json.Marshal(&q)
}

func (t *QueryGetLimitOrderTrancheResponse) MarshalBinding() ([]byte, error) {
	type QueryGetLimitOrderTrancheResponseBinding struct {
		LimitOrderTranche *LimitOrderTrancheBinding `json:"limit_order_tranche,omitempty"`
	}
	lo := t.LimitOrderTranche.ToBinding()
	q := QueryGetLimitOrderTrancheResponseBinding{
		LimitOrderTranche: &lo,
	}
	return json.Marshal(&q)
}

func (t *QueryAllInactiveLimitOrderTrancheResponse) MarshalBinding() ([]byte, error) {
	type QueryAllInactiveLimitOrderTrancheResponseBindings struct {
		InactiveLimitOrderTranche []*LimitOrderTrancheBinding `json:"inactive_limit_order_tranche,omitempty"`
		Pagination                *query.PageResponse         `json:"pagination,omitempty"`
	}
	q := QueryAllInactiveLimitOrderTrancheResponseBindings{
		InactiveLimitOrderTranche: make([]*LimitOrderTrancheBinding, 0, len(t.InactiveLimitOrderTranche)),
		Pagination:                t.Pagination,
	}
	for _, lo := range t.InactiveLimitOrderTranche {
		loNew := lo.ToBinding()
		q.InactiveLimitOrderTranche = append(q.InactiveLimitOrderTranche, &loNew)
	}

	return json.Marshal(&q)
}

func (t *QueryGetInactiveLimitOrderTrancheResponse) MarshalBinding() ([]byte, error) {
	type QueryGetInactiveLimitOrderTrancheResponseBinding struct {
		InactiveLimitOrderTranche *LimitOrderTrancheBinding `json:"inactive_limit_order_tranche,omitempty"`
	}
	lo := t.InactiveLimitOrderTranche.ToBinding()
	q := QueryGetInactiveLimitOrderTrancheResponseBinding{
		InactiveLimitOrderTranche: &lo,
	}
	return json.Marshal(&q)
}

func (t *QueryAllLimitOrderTrancheResponse) MarshalBinding() ([]byte, error) {
	type QueryAllLimitOrderTrancheResponseBinding struct {
		LimitOrderTranche []*LimitOrderTrancheBinding `json:"limit_order_tranche,omitempty"`
		Pagination        *query.PageResponse         `json:"pagination,omitempty"`
	}
	q := QueryAllLimitOrderTrancheResponseBinding{
		LimitOrderTranche: make([]*LimitOrderTrancheBinding, 0, len(t.LimitOrderTranche)),
		Pagination:        t.Pagination,
	}
	for _, lo := range t.LimitOrderTranche {
		loNew := lo.ToBinding()
		q.LimitOrderTranche = append(q.LimitOrderTranche, &loNew)
	}

	return json.Marshal(&q)
}

func (t *QueryGetLimitOrderTrancheUserResponse) MarshalBinding() ([]byte, error) {
	type QueryGetLimitOrderTrancheUserResponseBinding struct {
		LimitOrderTrancheUser *LimitOrderTrancheUserBinding `json:"limit_order_tranche_user,omitempty"`
	}

	lou := LimitOrderTrancheUserBinding{LimitOrderTrancheUser: *t.LimitOrderTrancheUser}
	s, ok := LimitOrderType_name[int32(t.LimitOrderTrancheUser.OrderType)]
	if !ok {
		return nil, errors.Wrap(ErrInvalidOrderType,
			fmt.Sprintf(
				"got \"%d\", expected one of %v",
				t.LimitOrderTrancheUser.OrderType,
				maps.Keys(LimitOrderType_name)),
		)
	}
	lou.OrderType = s
	q := QueryGetLimitOrderTrancheUserResponseBinding{LimitOrderTrancheUser: &lou}
	return json.Marshal(&q)
}

func (t *QueryAllLimitOrderTrancheUserResponse) MarshalBinding() ([]byte, error) {
	type QueryAllLimitOrderTrancheUserResponseBinding struct {
		LimitOrderTrancheUser []*LimitOrderTrancheUserBinding `json:"limit_order_tranche_user,omitempty"`
		Pagination            *query.PageResponse             `json:"pagination,omitempty"`
	}

	allLimitOrders := QueryAllLimitOrderTrancheUserResponseBinding{
		LimitOrderTrancheUser: make([]*LimitOrderTrancheUserBinding, 0, len(t.LimitOrderTrancheUser)),
		Pagination:            t.Pagination,
	}
	for _, lo := range t.LimitOrderTrancheUser {
		loNew := LimitOrderTrancheUserBinding{LimitOrderTrancheUser: *lo}
		s, ok := LimitOrderType_name[int32(lo.OrderType)]
		if !ok {
			return nil, errors.Wrap(ErrInvalidOrderType,
				fmt.Sprintf(
					"got \"%d\", expected one of %v",
					lo.OrderType,
					maps.Keys(LimitOrderType_name)),
			)
		}
		loNew.OrderType = s
		allLimitOrders.LimitOrderTrancheUser = append(allLimitOrders.LimitOrderTrancheUser, &loNew)
	}
	return json.Marshal(&allLimitOrders)
}

func (t *QueryAllLimitOrderTrancheUserByAddressResponse) MarshalBinding() ([]byte, error) {
	type QueryAllUserLimitOrdersResponseBinding struct {
		LimitOrders []*LimitOrderTrancheUserBinding `json:"limit_orders,omitempty"`
		Pagination  *query.PageResponse             `json:"pagination,omitempty"`
	}

	allLimitOrders := QueryAllUserLimitOrdersResponseBinding{
		LimitOrders: make([]*LimitOrderTrancheUserBinding, 0, len(t.LimitOrders)),
		Pagination:  t.Pagination,
	}
	for _, lo := range t.LimitOrders {
		loNew := LimitOrderTrancheUserBinding{LimitOrderTrancheUser: *lo}
		s, ok := LimitOrderType_name[int32(lo.OrderType)]
		if !ok {
			return nil, errors.Wrap(ErrInvalidOrderType,
				fmt.Sprintf(
					"got \"%d\", expected one of %v",
					lo.OrderType,
					maps.Keys(LimitOrderType_name)),
			)
		}
		loNew.OrderType = s
		allLimitOrders.LimitOrders = append(allLimitOrders.LimitOrders, &loNew)
	}
	return json.Marshal(&allLimitOrders)
}
