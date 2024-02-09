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

func (t LimitOrderTranche) ToBinding() LimitOrderTrancheBinding {
	lo := LimitOrderTrancheBinding{
		LimitOrderTranche: t,
	}
	if t.ExpirationTime != nil && t.ExpirationTime.Unix() >= 0 {
		ut := uint64(t.ExpirationTime.Unix())
		lo.ExpirationTime = &ut
	}
	return lo
}

type QueryGetPoolMetadataResponseBinding struct {
	PoolMetadata PoolMetadata `json:"pool_metadata"`
}

func (t *QueryGetPoolMetadataResponse) MarshalBinding() ([]byte, error) {
	b := QueryGetPoolMetadataResponseBinding(*t)
	return json.Marshal(&b)
}

func (t *QueryAllTickLiquidityResponse) MarshalBinding() ([]byte, error) {
	type liq struct {
		Liquidity struct {
			LimitOrderTranche *LimitOrderTrancheBinding `json:"limit_order_tranche,omitempty"`
			PoolReserves      *PoolReserves             `json:"pool_reserves,omitempty"`
		} `json:"liquidity,omitempty"`
	}
	type QueryAllTickLiquidityResponseBinding struct {
		TickLiquidity []liq               `json:"tick_liquidity,omitempty"`
		Pagination    *query.PageResponse `json:"pagination,omitempty"`
	}
	q := QueryAllTickLiquidityResponseBinding{
		TickLiquidity: make([]liq, 0, len(t.TickLiquidity)),
		Pagination:    t.Pagination,
	}
	for _, l := range t.TickLiquidity {
		if lq, ok := l.Liquidity.(*TickLiquidity_LimitOrderTranche); ok {
			lo := lq.LimitOrderTranche.ToBinding()
			l1 := liq{}
			l1.Liquidity.LimitOrderTranche = &lo
			q.TickLiquidity = append(q.TickLiquidity, l1)
		} else {
			l1 := liq{}
			l1.Liquidity.PoolReserves = l.Liquidity.(*TickLiquidity_PoolReserves).PoolReserves
			q.TickLiquidity = append(q.TickLiquidity, l1)
		}
	}
	return json.Marshal(&q)
}

func (t *QueryGetLimitOrderTrancheResponse) MarshalBinding() ([]byte, error) {
	type QueryGetLimitOrderTrancheResponseBinding struct {
		LimitOrderTranche *LimitOrderTrancheBinding `json:"limit_order_tranche,omitempty"`
	}
	lo := t.LimitOrderTranche.ToBinding()
	qr := QueryGetLimitOrderTrancheResponseBinding{
		LimitOrderTranche: &lo,
	}
	return json.Marshal(&qr)
}

func (t *QueryAllInactiveLimitOrderTrancheResponse) MarshalBinding() ([]byte, error) {
	type QueryAllInactiveLimitOrderTrancheResponseBindings struct {
		InactiveLimitOrderTranche []*LimitOrderTrancheBinding `json:"inactive_limit_order_tranche,omitempty"`
		Pagination                *query.PageResponse         `json:"pagination,omitempty"`
	}
	qr := QueryAllInactiveLimitOrderTrancheResponseBindings{
		InactiveLimitOrderTranche: make([]*LimitOrderTrancheBinding, 0, len(t.InactiveLimitOrderTranche)),
		Pagination:                t.Pagination,
	}
	for _, lo := range t.InactiveLimitOrderTranche {
		newlo := lo.ToBinding()
		qr.InactiveLimitOrderTranche = append(qr.InactiveLimitOrderTranche, &newlo)
	}

	return json.Marshal(&qr)
}

func (t *QueryGetInactiveLimitOrderTrancheResponse) MarshalBinding() ([]byte, error) {
	type QueryGetInactiveLimitOrderTrancheResponseBinding struct {
		InactiveLimitOrderTranche *LimitOrderTrancheBinding `json:"inactive_limit_order_tranche,omitempty"`
	}
	lo := t.InactiveLimitOrderTranche.ToBinding()
	qr := QueryGetInactiveLimitOrderTrancheResponseBinding{
		InactiveLimitOrderTranche: &lo,
	}
	return json.Marshal(&qr)
}

func (t *QueryAllLimitOrderTrancheResponse) MarshalBinding() ([]byte, error) {
	type QueryAllLimitOrderTrancheResponseBinding struct {
		LimitOrderTranche []*LimitOrderTrancheBinding `json:"limit_order_tranche,omitempty"`
		Pagination        *query.PageResponse         `json:"pagination,omitempty"`
	}
	qr := QueryAllLimitOrderTrancheResponseBinding{
		LimitOrderTranche: make([]*LimitOrderTrancheBinding, 0, len(t.LimitOrderTranche)),
		Pagination:        t.Pagination,
	}
	for _, lo := range t.LimitOrderTranche {
		newlo := lo.ToBinding()
		qr.LimitOrderTranche = append(qr.LimitOrderTranche, &newlo)
	}

	return json.Marshal(&qr)
}

type LimitOrderTrancheUserBinding struct {
	LimitOrderTrancheUser
	OrderType string `json:"order_type"`
}

type QueryGetLimitOrderTrancheUserResponseBinding struct {
	LimitOrderTrancheUser *LimitOrderTrancheUserBinding `json:"limit_order_tranche_user,omitempty"`
}

func (t *QueryGetLimitOrderTrancheUserResponse) MarshalBinding() ([]byte, error) {
	lou := LimitOrderTrancheUserBinding{LimitOrderTrancheUser: *t.LimitOrderTrancheUser}
	s, ok := LimitOrderType_name[int32(t.LimitOrderTrancheUser.OrderType)]
	if !ok {
		return nil, errors.Wrap(ErrInvalidOrderType,
			fmt.Sprintf(
				"got \"%d\", expeted one of %v",
				t.LimitOrderTrancheUser.OrderType,
				maps.Keys(LimitOrderType_name)),
		)
	}
	lou.OrderType = s
	r := QueryGetLimitOrderTrancheUserResponseBinding{LimitOrderTrancheUser: &lou}
	return json.Marshal(&r)
}

type QueryAllLimitOrderTrancheUserResponseBinding struct {
	LimitOrderTrancheUser []*LimitOrderTrancheUserBinding `json:"limit_order_tranche_user,omitempty"`
	Pagination            *query.PageResponse             `json:"pagination,omitempty"`
}

func (t *QueryAllLimitOrderTrancheUserResponse) MarshalBinding() ([]byte, error) {
	allLimitOrders := QueryAllLimitOrderTrancheUserResponseBinding{
		LimitOrderTrancheUser: make([]*LimitOrderTrancheUserBinding, 0, len(t.LimitOrderTrancheUser)),
		Pagination:            t.Pagination,
	}
	for _, lo := range t.LimitOrderTrancheUser {
		newlo := LimitOrderTrancheUserBinding{LimitOrderTrancheUser: *lo}
		s, ok := LimitOrderType_name[int32(lo.OrderType)]
		if !ok {
			return nil, errors.Wrap(ErrInvalidOrderType,
				fmt.Sprintf(
					"got \"%d\", expeted one of %v",
					lo.OrderType,
					maps.Keys(LimitOrderType_name)),
			)
		}
		newlo.OrderType = s
		allLimitOrders.LimitOrderTrancheUser = append(allLimitOrders.LimitOrderTrancheUser, &newlo)
	}
	return json.Marshal(&allLimitOrders)
}

type QueryAllUserLimitOrdersResponseBinding struct {
	LimitOrders []*LimitOrderTrancheUserBinding `json:"limit_orders,omitempty"`
	Pagination  *query.PageResponse             `json:"pagination,omitempty"`
}

func (t *QueryAllUserLimitOrdersResponse) MarshalBinding() ([]byte, error) {
	allLimitOrders := QueryAllUserLimitOrdersResponseBinding{
		LimitOrders: make([]*LimitOrderTrancheUserBinding, 0, len(t.LimitOrders)),
		Pagination:  t.Pagination,
	}
	for _, lo := range t.LimitOrders {
		newlo := LimitOrderTrancheUserBinding{LimitOrderTrancheUser: *lo}
		s, ok := LimitOrderType_name[int32(lo.OrderType)]
		if !ok {
			return nil, errors.Wrap(ErrInvalidOrderType,
				fmt.Sprintf(
					"got \"%d\", expeted one of %v",
					lo.OrderType,
					maps.Keys(LimitOrderType_name)),
			)
		}
		newlo.OrderType = s
		allLimitOrders.LimitOrders = append(allLimitOrders.LimitOrders, &newlo)
	}
	return json.Marshal(&allLimitOrders)
}
