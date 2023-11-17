package types

func NewPoolMetadata(pairID *PairID, tick int64, fee, poolID uint64) PoolMetadata {
	return PoolMetadata{PairId: pairID, Tick: tick, Fee: fee, Id: poolID}
}
