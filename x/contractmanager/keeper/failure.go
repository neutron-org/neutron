package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/neutron-org/neutron/v6/x/contractmanager/types"
)

// AddContractFailure adds a specific failure to the store. The provided address is used to determine
// the failure ID and they both are used to create a storage key for the failure.
//
// WARNING: The errMsg string parameter is expected to be deterministic. It means that the errMsg
// must be OS/library version agnostic and carry a concrete defined error message. One of the good
// ways to do so is to redact error using the RedactError func as it is done in SudoLimitWrapper
// Sudo method:
// https://github.com/neutron-org/neutron/blob/eb8b5ae50907439ff9af0527a42ef0cb448a78b5/x/contractmanager/ibc_middleware.go#L42.
// Another good way could be passing here some constant value.
func (k Keeper) AddContractFailure(ctx context.Context, address string, sudoPayload []byte, errMsg string) types.Failure {
	c := sdk.UnwrapSDKContext(ctx)

	failure := types.Failure{
		Address:     address,
		SudoPayload: sudoPayload,
		Error:       errMsg,
	}
	nextFailureID := k.GetNextFailureIDKey(ctx, failure.GetAddress())
	failure.Id = nextFailureID

	store := c.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&failure)
	store.Set(types.GetFailureKey(failure.GetAddress(), nextFailureID), bz)
	return failure
}

func (k Keeper) GetNextFailureIDKey(ctx context.Context, address string) uint64 {
	c := sdk.UnwrapSDKContext(ctx)

	store := prefix.NewStore(c.KVStore(k.storeKey), types.GetFailureKeyPrefix(address))
	iterator := storetypes.KVStoreReversePrefixIterator(store, []byte{})
	defer iterator.Close()

	if iterator.Valid() {
		var val types.Failure
		k.cdc.MustUnmarshal(iterator.Value(), &val)

		return val.Id + 1
	}

	return 0
}

// GetAllFailures returns all failures
func (k Keeper) GetAllFailures(ctx context.Context) (list []types.Failure) {
	c := sdk.UnwrapSDKContext(ctx)

	store := prefix.NewStore(c.KVStore(k.storeKey), types.ContractFailuresKey)
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Failure
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

func (k Keeper) GetFailure(ctx sdk.Context, contractAddr sdk.AccAddress, id uint64) (*types.Failure, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetFailureKey(contractAddr.String(), id)

	bz := store.Get(key)
	if bz == nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "no failure found for contractAddress = %s and failureId = %d", contractAddr.String(), id)
	}
	var res types.Failure
	k.cdc.MustUnmarshal(bz, &res)

	return &res, nil
}

// resubmitFailure tries to call sudo handler for contract with same parameters as initially.
// if successful, removes the failure from storage
func (k Keeper) resubmitFailure(ctx sdk.Context, contractAddr sdk.AccAddress, failure *types.Failure) error {
	if failure.SudoPayload == nil {
		return errorsmod.Wrapf(types.ErrIncorrectFailureToResubmit, "cannot resubmit failure without sudo payload; failureId = %d", failure.Id)
	}

	if _, err := k.wasmKeeper.Sudo(ctx, contractAddr, failure.SudoPayload); err != nil {
		return errorsmod.Wrapf(types.ErrFailedToResubmitFailure, "cannot resubmit failure; failureId = %d; err = %s", failure.Id, err)
	}

	// Cleanup failure since we resubmitted it successfully
	k.removeFailure(ctx, contractAddr, failure.Id)

	return nil
}

func (k Keeper) removeFailure(ctx sdk.Context, contractAddr sdk.AccAddress, id uint64) {
	store := ctx.KVStore(k.storeKey)
	failureKey := types.GetFailureKey(contractAddr.String(), id)
	store.Delete(failureKey)
}

// RedactError removes non-determenistic details from the error returning just codespace and core
// of the error. Returns full error for system errors.
//
// Copy+paste from https://github.com/neutron-org/wasmd/blob/5b59886e41ed55a7a4a9ae196e34b0852285503d/x/wasm/keeper/msg_dispatcher.go#L175-L190
func RedactError(err error) error {
	// Do not redact system errors
	// SystemErrors must be created in x/wasm and we can ensure determinism
	if wasmvmtypes.ToSystemError(err) != nil {
		return err
	}

	// FIXME: do we want to hardcode some constant string mappings here as well?
	// Or better document them? (SDK error string may change on a patch release to fix wording)
	// sdk/11 is out of gas
	// sdk/5 is insufficient funds (on bank send)
	// (we can theoretically redact less in the future, but this is a first step to safety)
	codespace, code, _ := errorsmod.ABCIInfo(err, false)
	return fmt.Errorf("codespace: %s, code: %d", codespace, code)
}
