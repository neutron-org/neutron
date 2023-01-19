package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type PayerInfo struct {
	Sender   sdk.AccAddress
	FeePayer sdk.AccAddress
}
