package types

import (
	math_utils "github.com/neutron-org/neutron/v10/utils/math"
)

func (l LimitOrderTrancheUser) IsEmpty() bool {
	return l.DecSharesWithdrawn.Equal(math_utils.NewPrecDecFromInt(l.SharesOwned))
}

func (l *LimitOrderTrancheUser) SetSharesWithdrawn(shares math_utils.PrecDec) {
	l.SharesWithdrawn = shares.TruncateInt()
	l.DecSharesWithdrawn = shares
}
