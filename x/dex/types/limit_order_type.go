package types

func (l LimitOrderType) IsGTC() bool {
	return l == LimitOrderType_GOOD_TIL_CANCELLED
}

func (l LimitOrderType) IsFoK() bool {
	return l == LimitOrderType_FILL_OR_KILL
}

func (l LimitOrderType) IsIoC() bool {
	return l == LimitOrderType_IMMEDIATE_OR_CANCEL
}

func (l LimitOrderType) IsJIT() bool {
	return l == LimitOrderType_JUST_IN_TIME
}

func (l LimitOrderType) IsGoodTil() bool {
	return l == LimitOrderType_GOOD_TIL_TIME
}

func (l LimitOrderType) IsTakerOnly() bool {
	return l.IsIoC() || l.IsFoK()
}

func (l LimitOrderType) HasExpiration() bool {
	return l.IsGoodTil() || l.IsJIT()
}
