package types

func NewPoolMetadata(pairID *PairID, tick int64, fee uint64, poolID uint64) PoolMetadata {
	return PoolMetadata{PairID: pairID, Tick: tick, Fee: fee, ID: poolID}
}
