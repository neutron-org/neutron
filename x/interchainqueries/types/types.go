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
	AttributeKVQueryKey              = "kv_key"
	AttributeTransactionsFilterQuery = "tx_filter"

	AttributeValueCategory = ModuleName
	AttributeValueQuery    = "query"
)

const (
	InterchainQueryTypeKV = "kv"
	InterchainQueryTypeTX = "tx"

	delimiter = "/"
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
	return kv.Path + delimiter + hex.EncodeToString(kv.Key)
}

func KVKeyFromString(s string) (KVKey, error) {
	splittedString := strings.Split(s, delimiter)
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
