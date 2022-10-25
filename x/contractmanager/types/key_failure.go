package types

import "encoding/binary"

var _ binary.ByteOrder

const (
    // FailureKeyPrefix is the prefix to retrieve all Failure
	FailureKeyPrefix = "Failure/value/"
)

// FailureKey returns the store key to retrieve a Failure from the index fields
func FailureKey(
index string,
) []byte {
	var key []byte
    
    indexBytes := []byte(index)
    key = append(key, indexBytes...)
    key = append(key, []byte("/")...)
    
	return key
}