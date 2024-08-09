package types

func (l LimitOrderTrancheUser) IsEmpty() bool {
	return l.SharesWithdrawn.Equal(l.SharesOwned)
}
