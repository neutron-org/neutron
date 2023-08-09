package keeper

import (
	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/neutron-org/neutron/x/contractmanager/types"
)

// AddContractFailure adds a specific failure to the store using address as the key
func (k Keeper) AddContractFailure(ctx sdk.Context, packet ibcchanneltypes.Packet, address string, ackType string, ackResult []byte, errorText string) {
	failure := types.Failure{
		ChannelId: packet.SourceChannel,
		Address:   address,
		AckId:     packet.Sequence,
		AckType:   ackType,
		AckResult: ackResult,
		ErrorText: errorText,
		Packet:    &packet,
	}
	nextFailureID := k.GetNextFailureIDKey(ctx, failure.GetAddress())

	store := ctx.KVStore(k.storeKey)

	failure.Id = nextFailureID
	b := k.cdc.MustMarshal(&failure)
	store.Set(types.GetFailureKey(failure.GetAddress(), nextFailureID), b)
}

func (k Keeper) GetNextFailureIDKey(ctx sdk.Context, address string) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetFailureKeyPrefix(address))
	iterator := sdk.KVStoreReversePrefixIterator(store, []byte{})
	defer iterator.Close()

	if iterator.Valid() {
		var val types.Failure
		k.cdc.MustUnmarshal(iterator.Value(), &val)

		return val.Id + 1
	}

	return 0
}

// GetAllFailure returns all failures
func (k Keeper) GetAllFailures(ctx sdk.Context) (list []types.Failure) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ContractFailuresKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Failure
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

func (k Keeper) GetFailure(ctx sdk.Context, contractAddr sdk.AccAddress, failureId uint64) (*types.Failure, error) {
	store := ctx.KVStore(k.storeKey)
	failureKey := types.GetFailureKey(contractAddr.String(), failureId)

	failureBz := store.Get(failureKey)
	if failureBz == nil {
		return nil, errors.Wrapf(sdkerrors.ErrKeyNotFound, "no failure found for contractAddress = %s and failureId = %d", contractAddr.String(), failureId)
	}
	var failure types.Failure
	k.cdc.MustUnmarshal(failureBz, &failure)

	return &failure, nil
}

// ResubmitFailure tries to call sudo handler for contract with same parameters as initially.
func (k Keeper) ResubmitFailure(ctx sdk.Context, contractAddr sdk.AccAddress, failure *types.Failure) error {
	if failure.Packet == nil {
		return errors.Wrapf(types.IncorrectFailureToResubmit, "cannot resubmit failure without packet info failureId = %d", failure.Id)
	}

	if failure.GetAckType() == "ack" { // response or error
		_, err := k.SudoResponse(ctx, contractAddr, *failure.Packet, failure.AckResult)
		// TODO: handle resp?
		if err != nil {
			return err // TODO: wrap
		}
	} else if failure.GetAckType() == "timeout" {
		// TODO
		_, err := k.SudoTimeout(ctx, contractAddr, *failure.Packet)
		if err != nil {
			return err // TODO: wrap
		}
	} else {
		return errors.Wrapf(types.IncorrectAckType, "cannot resubmit failure with incorrect ackType = %s", failure.GetAckType())
	}

	// TODO: If submitted failure response or timeout successfully, we can cleanup it?
	// Or maybe mark it as processed
	// Also maybe cleanup packet and ack data to smaller data to store
	// Is it bad be able to call resubmitFailure multiple times?
	k.removeFailure(ctx, contractAddr, failure.Id)

	// TODO: maybe return result from sudo call?

	return nil
}

func (k Keeper) removeFailure(ctx sdk.Context, contractAddr sdk.AccAddress, id uint64) {
	store := ctx.KVStore(k.storeKey)
	failureKey := types.GetFailureKey(contractAddr.String(), id)
	store.Delete(failureKey)
}
