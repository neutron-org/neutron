package types

import (
	"encoding/hex"
	"encoding/json"
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// AttributeKeyQueryID represents the key for event attribute delivering the query ID of a
	// registered interchain query.
	AttributeKeyQueryID = "query_id"

	// AttributeKeyOwner represents the key for event attribute delivering the address of the
	// registrator of an interchain query.
	AttributeKeyOwner = "owner"

	// AttributeKeyZoneID represents the key for event attribute delivering the zone ID where the
	// event has been produced.
	AttributeKeyZoneID = "zone_id"

	// AttributeKeyQueryType represents the key for event attribute delivering the query type
	// identifier (e.g. 'kv' or 'tx')
	AttributeKeyQueryType = "type"

	// AttributeKeyKVQuery represents the keys of the storage we want to get from remote chain for event attribute delivering the keys
	// of an interchain query.
	AttributeKeyKVQuery = "kv_key"

	// AttributeTransactionsFilterQuery represents the transactions filter for event attribute delivering the filter
	// of an interchain query.
	AttributeTransactionsFilterQuery = "tx_filter"

	// AttributeValueCategory represents the value for the 'module' event attribute.
	AttributeValueCategory = ModuleName
	// AttributeValueQuery represents the value for the 'action' event attribute.
	AttributeValueQuery = "query"
)

const (
	InterchainQueryTypeKV = "kv"
	InterchainQueryTypeTX = "tx"

	kvPathKeyDelimiter = "/"
	kvKeysDelimiter    = ","
)

type FilterItem struct {
	Field string
	Op    string
	Value interface{}
}

func IsValidTransactionFilterJSON(s string) bool {
	var js []FilterItem
	return json.Unmarshal([]byte(s), &js) == nil

}

type InterchainQueryType string

func (icqt InterchainQueryType) IsValid() bool {
	return icqt.IsTX() || icqt.IsKV()
}

func (icqt InterchainQueryType) IsKV() bool {
	return icqt == InterchainQueryTypeKV
}

func (icqt InterchainQueryType) IsTX() bool {
	return icqt == InterchainQueryTypeTX
}

func (kv KVKey) ToString() string {
	return kv.Path + kvPathKeyDelimiter + hex.EncodeToString(kv.Key)
}

func KVKeyFromString(s string) (KVKey, error) {
	splitString := strings.Split(s, kvPathKeyDelimiter)
	if len(splitString) < 2 {
		return KVKey{}, sdkerrors.Wrap(ErrInvalidType, "invalid kv key type")
	}

	bzKey, err := hex.DecodeString(splitString[1])
	if err != nil {
		return KVKey{}, sdkerrors.Wrapf(err, "invalid key encoding")
	}
	return KVKey{
		Path: splitString[0],
		Key:  bzKey,
	}, nil
}

type KVKeys []*KVKey

func KVKeysFromString(str string) (KVKeys, error) {
	splitString := strings.Split(str, kvKeysDelimiter)
	kvKeys := make(KVKeys, 0, len(splitString))

	for _, s := range splitString {
		key, err := KVKeyFromString(s)
		if err != nil {
			return nil, err
		}
		kvKeys = append(kvKeys, &key)
	}

	return kvKeys, nil
}

func (keys KVKeys) String() string {
	if len(keys) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(keys[0].ToString())

	for _, key := range keys[1:] {
		b.WriteString(kvKeysDelimiter)
		b.WriteString(key.ToString())
	}

	return b.String()
}
