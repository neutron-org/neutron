package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	math_utils "github.com/neutron-org/neutron/utils/math"
)

const TypeMsgMultiHopSwap = "multi_hop_swap"

var _ sdk.Msg = &MsgMultiHopSwap{}

func NewMsgMultiHopSwap(
	creator string,
	receiever string,
	routesArr [][]string,
	amountIn sdk.Int,
	exitLimitPrice math_utils.PrecDec,
	pickBestRoute bool,
) *MsgMultiHopSwap {
	routes := make([]*MultiHopRoute, len(routesArr))
	for i, hops := range routesArr {
		routes[i] = &MultiHopRoute{Hops: hops}
	}

	return &MsgMultiHopSwap{
		Creator:        creator,
		Receiver:       receiever,
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
	return sdk.MustSortJSON(bz)
}

func (msg *MsgMultiHopSwap) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	_, err = sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid receiver address (%s)", err)
	}

	if len(msg.Routes) == 0 {
		return ErrMissingMultihopRoute
	}

	expectedExitToken := msg.Routes[0].Hops[len(msg.Routes[0].Hops)-1]
	for _, route := range msg.Routes[1:] {
		hops := route.Hops
		if expectedExitToken != hops[len(hops)-1] {
			return ErrMultihopExitTokensMismatch
		}
	}

	if msg.AmountIn.LTE(sdk.ZeroInt()) {
		return ErrZeroSwap
	}

	return nil
}
