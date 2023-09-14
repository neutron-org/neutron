package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Sudo interface {
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
}
