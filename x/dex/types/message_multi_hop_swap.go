package types

import (
	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	math_utils "github.com/neutron-org/neutron/v6/utils/math"
)

const TypeMsgMultiHopSwap = "multi_hop_swap"

var _ sdk.Msg = &MsgMultiHopSwap{}

func NewMsgMultiHopSwap(
	creator string,
	receiver string,
	routesArr [][]string,
	amountIn math.Int,
	exitLimitPrice math_utils.PrecDec,
	pickBestRoute bool,
) *MsgMultiHopSwap {
	routes := make([]*MultiHopRoute, len(routesArr))
	for i, hops := range routesArr {
		routes[i] = &MultiHopRoute{Hops: hops}
	}

	return &MsgMultiHopSwap{
		Creator:        creator,
		Receiver:       receiver,
		Routes:         routes,
		AmountIn:       amountIn,
		ExitLimitPrice: exitLimitPrice,
		PickBestRoute:  pickBestRoute,
	}
}

func (msg *MsgMultiHopSwap) Route() string {
	return RouterKey
}

func (msg *MsgMultiHopSwap) Type() string {
	return TypeMsgMultiHopSwap
}

func (msg *MsgMultiHopSwap) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{creator}
}

func (msg *MsgMultiHopSwap) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return bz
}

func (msg *MsgMultiHopSwap) Validate() error {
	if err := validateAddress(msg.Creator, "creator"); err != nil {
		return err
	}
	if err := validateAddress(msg.Receiver, "receiver"); err != nil {
		return err
	}
	if err := validateRoutes(msg.Routes); err != nil {
		return err
	}
	if err := validateAmountIn(msg.AmountIn); err != nil {
		return err
	}
	if err := validateExitLimitPrice(msg.ExitLimitPrice); err != nil {
		return err
	}
	return nil
}

func validateAddress(address, field string) error {
	_, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidAddress, "invalid %s address (%s)", field, err)
	}
	return nil
}

func validateRoutes(routes []*MultiHopRoute) error {
	if len(routes) == 0 {
		return ErrMissingMultihopRoute
	}
	expectedExitToken := ""
	expectedEntryToken := ""

	for i, route := range routes {
		if err := validateHops(route.Hops); err != nil {
			return err
		}

		// validateHops ensures hops[] is at least length 2
		exitToken := route.Hops[len(route.Hops)-1]
		entryToken := route.Hops[0]

		switch {
		case i == 0:
			expectedExitToken = exitToken
			expectedEntryToken = entryToken
		case exitToken != expectedExitToken:
			return ErrMultihopExitTokensMismatch
		case entryToken != expectedEntryToken:
			return ErrMultihopEntryTokensMismatch
		}
	}
	return nil
}

func validateHops(hops []string) error {
	existingHops := make(map[string]bool, len(hops))
	for _, hop := range hops {
		// check that route has at least entry and exit token
		if len(hop) < 2 {
			return ErrRouteWithoutExitToken
		}
		// check if we find cycles in the route
		if existingHops[hop] {
			return ErrCycleInHops
		}
		existingHops[hop] = true

		// check if the denom is valid
		err := sdk.ValidateDenom(hop)
		if err != nil {
			return sdkerrors.Wrapf(ErrInvalidDenom, "invalid denom in route: (%s)", err)
		}
	}
	return nil
}

func validateAmountIn(amount math.Int) error {
	if amount.LTE(math.ZeroInt()) {
		return ErrZeroSwap
	}
	return nil
}

func validateExitLimitPrice(price math_utils.PrecDec) error {
	if !price.IsPositive() {
		return ErrZeroExitPrice
	}
	return nil
}
