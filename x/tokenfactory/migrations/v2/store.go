package v2

import (
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/store/prefix"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/tokenfactory/types"
	v1beta1types "github.com/neutron-org/neutron/v6/x/tokenfactory/types/v1beta1"
)

type TokenFactoryKeeper interface {
	AssertIsHookWhitelisted(ctx sdk.Context, denom string, contractAddress sdk.AccAddress) error
}

// MigrateStore performs in-place store migrations.
// The migration adds the new tokenfactory params WhitelistedHooks
func MigrateStore(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey, keeper TokenFactoryKeeper) error {
	// NOTE: this must happen first since the migrateHooks relies on the new params
	if err := migrateParams(ctx, cdc, storeKey); err != nil {
		return err
	}
	if err := migrateHooks(ctx, storeKey, keeper); err != nil {
		return err
	}

	return nil
}

var WhitelistedHooks = []*types.WhitelistedHook{
	{ // xASTRO balances tracker
		CodeID:       944,
		DenomCreator: "neutron1zlf3hutsa4qnmue53lz2tfxrutp8y2e3rj4nkghg3rupgl4mqy8s5jgxsn",
	},
	{ // USDC <> NTRN balances tracker
		CodeID:       1473,
		DenomCreator: "neutron18c8qejysp4hgcfuxdpj4wf29mevzwllz5yh8uayjxamwtrs0n9fshq9vtv",
	},
	{ // NTRN <> ATOM balances tracker
		CodeID:       1473,
		DenomCreator: "neutron145z3nj7yqft2vpugr5a5p7jsnagvms90tvtej45g4s0xkqalhy7sj20vgz",
	},
	{ // NFA.zoneV1
		CodeID:       1265,
		DenomCreator: "neutron1pwjn3tsumm3j7v7clzqhjsaukv4tdjlclhdytawhet68fwlz84fqcrdyf5",
	},
}

func migrateParams(ctx sdk.Context, cdc codec.BinaryCodec, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Migrating tokenfactory params...")

	// fetch old params
	store := ctx.KVStore(storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return errors.New("cannot fetch tokenfactory params from KV store")
	}
	var oldParams v1beta1types.Params
	cdc.MustUnmarshal(bz, &oldParams)

	// add new param values
	newParams := types.Params{
		DenomCreationFee:        oldParams.DenomCreationFee,
		DenomCreationGasConsume: oldParams.DenomCreationGasConsume,
		FeeCollectorAddress:     oldParams.FeeCollectorAddress,
		WhitelistedHooks:        WhitelistedHooks,
	}

	// set params
	bz, err := cdc.Marshal(&newParams)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)

	ctx.Logger().Info("Finished migrating tokenfactory params")

	return nil
}

func migrateHooks(ctx sdk.Context, storeKey storetypes.StoreKey, keeper TokenFactoryKeeper) error {
	ctx.Logger().Info("Migrating tokenfactory hooks...")

	// get hook store
	store := prefix.NewStore(ctx.KVStore(storeKey), []byte(types.DenomsPrefixKey))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	hooksToRemove := make([][]byte, 0)
	for ; iterator.Valid(); iterator.Next() {
		keyParts := strings.Split(string(iterator.Key()), types.KeySeparator)
		if len(keyParts) != 3 {
			return fmt.Errorf("cannot parse BeforeSendHook data key: %s", string(iterator.Key()))
		}

		// Hooks and authorityMetadata are in the same store, we only care about the hooks
		if keyParts[2] == types.BeforeSendHookAddressPrefixKey {
			denom := keyParts[1]
			contractAddr, err := sdk.AccAddressFromBech32(string(iterator.Value()))
			if err != nil {
				return fmt.Errorf("cannot parse hook contract address: %s", string(iterator.Value()))
			}

			err = keeper.AssertIsHookWhitelisted(ctx, denom, contractAddr)
			if err != nil {
				hooksToRemove = append(hooksToRemove, iterator.Key())
			}
		}
	}

	err := iterator.Close()
	if err != nil {
		return errorsmod.Wrap(err, "iterator failed to close after migration")
	}

	// Delete all non-whitelisted hooks
	for _, k := range hooksToRemove {
		store.Delete(k)
	}

	ctx.Logger().Info("Finished migrating tokenfactory hooks")

	return nil
}
