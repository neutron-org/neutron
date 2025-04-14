package v3

import (
	store "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/neutron-org/neutron/v6/x/globalfee/types"
)

// MigrateStore performs in-place params migrations
// params module store the module store
func MigrateStore(ctx sdk.Context, cdc codec.BinaryCodec, globalfeeSubspace paramtypes.Subspace, storeKey store.StoreKey) error {
	params := types.Params{}
	globalfeeSubspace.GetParamSet(ctx, &params)
	store := ctx.KVStore(storeKey)
	bz := cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)
	return nil
}
