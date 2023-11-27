package types

func (l LimitOrderTrancheUser) IsEmpty() bool {
	sharesRemoved := l.SharesCancelled.Add(l.SharesWithdrawn)
	return sharesRemoved.Equal(l.SharesOwned)
}
