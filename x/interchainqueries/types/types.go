package types

import (
	"encoding/hex"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"strings"
)

const (
	AttributeKeyQueryID              = "query_id"
	AttributeKeyZoneID               = "zone_id"
	AttributeQueryType               = "type"
	AttributeKeyKVQuery              = "kv_key"
	AttributeTransactionsFilterQuery = "tx_filter"

	AttributeValueCategory = ModuleName
	AttributeValueQuery    = "query"
)

const (
	InterchainQueryTypeKV = "kv"
	InterchainQueryTypeTX = "tx"

	pathKeyDelimiter = "/"
	kvKeysDelimiter  = ","
)

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
	return kv.Path + pathKeyDelimiter + hex.EncodeToString(kv.Key)
}

func KVKeyFromString(s string) (KVKey, error) {
	splittedString := strings.Split(s, pathKeyDelimiter)
	if len(splittedString) < 2 {
		return KVKey{}, sdkerrors.Wrap(ErrInvalidType, "invalid kv key type")
	}

	bzKey, err := hex.DecodeString(splittedString[1])
	if err != nil {
		return KVKey{}, sdkerrors.Wrapf(err, "invalid key encoding")
	}
	return KVKey{
		Path: splittedString[0],
		Key:  bzKey,
	}, nil
}

type KVKeys []*KVKey

func KVKeysFromString(str string) (KVKeys, error) {
	splittedString := strings.Split(str, kvKeysDelimiter)
	kvKeys := make(KVKeys, 0, len(splittedString))

	for _, s := range splittedString {
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
