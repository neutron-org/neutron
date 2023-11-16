package types

import (
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/query"
	math_utils "github.com/neutron-org/neutron/utils/math"
)

/*
File contains an implementation of query responses custom json marshaller for wasmbinding
*/

import (
	"encoding/json"
)

type LimitOrderTrancheBinding struct {
	Key                *LimitOrderTrancheKey `json:"key,omitempty"`
	ReservesMakerDenom math.Int              `json:"reserves_maker_denom" yaml:"reserves_maker_denom"`
	ReservesTakerDenom math.Int              `json:"reserves_taker_denom" yaml:"reserves_taker_denom"`
	TotalMakerDenom    math.Int              `json:"total_maker_denom" yaml:"total_maker_denom"`
	TotalTakerDenom    math.Int              `json:"total_taker_denom" yaml:"total_taker_denom"`
	// JIT orders also use goodTilDate to handle deletion but represent a special case
	// All JIT orders have a goodTilDate of 0 and an exception is made to still still treat these orders as live
	// Order deletion still functions the same and the orders will be deleted at the end of the block
	ExpirationTime    *uint64            `json:"expiration_time,omitempty"`
	PriceTakerToMaker math_utils.PrecDec `json:"price_taker_to_maker" yaml:"price_taker_to_maker"`
}

func (t *QueryGetLimitOrderTrancheResponse) MarshalBinding() ([]byte, error) {
	type QueryGetLimitOrderTrancheResponseBinding struct {
		LimitOrderTranche *LimitOrderTrancheBinding `json:"limit_order_tranche,omitempty"`
	}
	lo := LimitOrderTrancheBinding{
		Key:                t.LimitOrderTranche.Key,
		ReservesMakerDenom: t.LimitOrderTranche.ReservesMakerDenom,
		ReservesTakerDenom: t.LimitOrderTranche.ReservesTakerDenom,
		TotalMakerDenom:    t.LimitOrderTranche.TotalMakerDenom,
		TotalTakerDenom:    t.LimitOrderTranche.TotalTakerDenom,
		PriceTakerToMaker:  t.LimitOrderTranche.PriceTakerToMaker,
	}
	if t.LimitOrderTranche.ExpirationTime != nil && !t.LimitOrderTranche.ExpirationTime.IsZero() {
		ut := uint64(t.LimitOrderTranche.ExpirationTime.Unix())
		lo.ExpirationTime = &ut
	}
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
		newlo := LimitOrderTrancheBinding{
			Key:                lo.Key,
			ReservesMakerDenom: lo.ReservesMakerDenom,
			ReservesTakerDenom: lo.ReservesTakerDenom,
			TotalMakerDenom:    lo.TotalMakerDenom,
			TotalTakerDenom:    lo.TotalTakerDenom,
			PriceTakerToMaker:  lo.PriceTakerToMaker,
		}
		if lo.ExpirationTime != nil && !lo.ExpirationTime.IsZero() {
			ut := uint64(lo.ExpirationTime.Unix())
			newlo.ExpirationTime = &ut
		}
		qr.InactiveLimitOrderTranche = append(qr.InactiveLimitOrderTranche, &newlo)
	}

	return json.Marshal(&qr)
}

func (t *QueryGetInactiveLimitOrderTrancheResponse) MarshalBinding() ([]byte, error) {
	type QueryGetInactiveLimitOrderTrancheResponseBinding struct {
		InactiveLimitOrderTranche *LimitOrderTrancheBinding `json:"inactive_limit_order_tranche,omitempty"`
	}
	lo := LimitOrderTrancheBinding{
		Key:                t.InactiveLimitOrderTranche.Key,
		ReservesMakerDenom: t.InactiveLimitOrderTranche.ReservesMakerDenom,
		ReservesTakerDenom: t.InactiveLimitOrderTranche.ReservesTakerDenom,
		TotalMakerDenom:    t.InactiveLimitOrderTranche.TotalMakerDenom,
		TotalTakerDenom:    t.InactiveLimitOrderTranche.TotalTakerDenom,
		PriceTakerToMaker:  t.InactiveLimitOrderTranche.PriceTakerToMaker,
	}
	if t.InactiveLimitOrderTranche.ExpirationTime != nil && !t.InactiveLimitOrderTranche.ExpirationTime.IsZero() {
		ut := uint64(t.InactiveLimitOrderTranche.ExpirationTime.Unix())
		lo.ExpirationTime = &ut
	}
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
		newlo := LimitOrderTrancheBinding{
			Key:                lo.Key,
			ReservesMakerDenom: lo.ReservesMakerDenom,
			ReservesTakerDenom: lo.ReservesTakerDenom,
			TotalMakerDenom:    lo.TotalMakerDenom,
			TotalTakerDenom:    lo.TotalTakerDenom,
			PriceTakerToMaker:  lo.PriceTakerToMaker,
		}
		if lo.ExpirationTime != nil && !lo.ExpirationTime.IsZero() {
			ut := uint64(lo.ExpirationTime.Unix())
			newlo.ExpirationTime = &ut
		}
		qr.LimitOrderTranche = append(qr.LimitOrderTranche, &newlo)
	}

	return json.Marshal(&qr)
}
