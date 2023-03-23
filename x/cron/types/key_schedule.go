package types

import "encoding/binary"

var _ binary.ByteOrder

const (
    // ScheduleKeyPrefix is the prefix to retrieve all Schedule
	ScheduleKeyPrefix = "Schedule/value/"
)

// ScheduleKey returns the store key to retrieve a Schedule from the index fields
func ScheduleKey(
index string,
) []byte {
	var key []byte
    
    indexBytes := []byte(index)
    key = append(key, indexBytes...)
    key = append(key, []byte("/")...)
    
	return key
}