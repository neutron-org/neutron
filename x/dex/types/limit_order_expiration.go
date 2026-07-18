package types

// Creates a new LimitOrderExpiration struct based on a LimitOrderTranche
func NewLimitOrderExpiration(tranche *LimitOrderTranche) *LimitOrderExpiration {
	trancheExpiry := tranche.ExpirationTime
	if trancheExpiry == nil {
		panic("Cannot create LimitOrderExpiration from tranche with nil ExpirationTime")
	}

	return &LimitOrderExpiration{
		TrancheRef:     tranche.Key.KeyMarshal(),
		ExpirationTime: *tranche.ExpirationTime,
	}
}
