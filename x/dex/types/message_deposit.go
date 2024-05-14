package types

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgDeposit = "deposit"

var _ sdk.Msg = &MsgDeposit{}

func NewMsgDeposit(
	creator,
	receiver,
	tokenA,
	tokenB string,
	amountsA,
	amountsB []math.Int,
	tickIndexes []int64,
	fees []uint64,
	depositOptions []*DepositOptions,
) *MsgDeposit {
	return &MsgDeposit{
		Creator:         creator,
		Receiver:        receiver,
		TokenA:          tokenA,
		TokenB:          tokenB,
		AmountsA:        amountsA,
		AmountsB:        amountsB,
		TickIndexesAToB: tickIndexes,
		Fees:            fees,
		Options:         depositOptions,
	}
}

func (msg *MsgDeposit) Route() string {
	return RouterKey
}

func (msg *MsgDeposit) Type() string {
	return TypeMsgDeposit
}

func (msg *MsgDeposit) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{creator}
}

func (msg *MsgDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return bz
}

func (msg *MsgDeposit) Validate() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Receiver)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidAddress, "invalid receiver address (%s)", err)
	}

	// Verify that the lengths of TickIndexes, Fees, AmountsA, AmountsB are all equal
	numDeposits := len(msg.AmountsA)
	if numDeposits != len(msg.Fees) ||
		numDeposits != len(msg.TickIndexesAToB) ||
		numDeposits != len(msg.AmountsB) ||
		numDeposits != len(msg.Options) {
		return ErrUnbalancedTxArray
	}
	if numDeposits == 0 {
		return ErrZeroDeposit
	}

	poolsDeposited := make(map[string]bool)
	for i := 0; i < numDeposits; i++ {
		poolStr := fmt.Sprintf("%d-%d", msg.TickIndexesAToB[i], msg.Fees[i])
		if _, ok := poolsDeposited[poolStr]; ok {
			return ErrDuplicatePoolDeposit
		}
		poolsDeposited[poolStr] = true
		if msg.AmountsA[i].LT(math.ZeroInt()) || msg.AmountsB[i].LT(math.ZeroInt()) {
			return ErrZeroDeposit
		}
		if msg.AmountsA[i].Equal(math.ZeroInt()) && msg.AmountsB[i].Equal(math.ZeroInt()) {
			return ErrZeroDeposit
		}
		if err := ValidateTickFee(msg.TickIndexesAToB[i], msg.Fees[i]); err != nil {
			return err
		}
	}

	return nil
}
