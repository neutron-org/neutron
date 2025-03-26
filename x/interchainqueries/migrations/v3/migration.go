package v3

import (
	"fmt"

	store "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/interchainqueries/types"
)

func MigrateParams(ctx sdk.Context, cdc codec.BinaryCodec, storeKey store.StoreKey) error {
	var params types.Params
	st := ctx.KVStore(storeKey)
	bz := st.Get(types.ParamsKey)
	if bz == nil {
		return fmt.Errorf("no params stored in %s", types.ParamsKey)
	}

	cdc.MustUnmarshal(bz, &params)
	params.MaxTransactionsFilters = types.DefaultMaxTransactionsFilters
	params.MaxKvQueryKeysCount = types.DefaultMaxKvQueryKeysCount
	bz = cdc.MustMarshal(&params)
	st.Set(types.ParamsKey, bz)
	return nil
}
