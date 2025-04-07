package keeper

import (
	"fmt"

	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/header"
	"cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibccommitmenttypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	tendermint "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"

	"github.com/neutron-org/neutron/v6/utils/storageverification"
	icqtypes "github.com/neutron-org/neutron/v6/x/interchainqueries/types"
	"github.com/neutron-org/neutron/v6/x/state-verifier/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		cometInfo  comet.BlockInfoService
		headerInfo header.Service
		authority  string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	cometInfo comet.BlockInfoService,
	headerInfo header.Service,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		authority:  authority,
		headerInfo: headerInfo,
		cometInfo:  cometInfo,
	}
}

func (k *Keeper) GetAuthority() string {
	return k.authority
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SaveConsensusState extracts info about the current header from the context, composes ConsensusState structure with that info
// and saves the structure to the state
func (k *Keeper) SaveConsensusState(ctx sdk.Context) error {
	headerInfo := k.headerInfo.GetHeaderInfo(ctx)
	cometInfo := k.cometInfo.GetCometBlockInfo(ctx)

	cs := tendermint.ConsensusState{
		Timestamp:          ctx.BlockTime(),
		Root:               ibccommitmenttypes.NewMerkleRoot(headerInfo.AppHash),
		NextValidatorsHash: cometInfo.GetValidatorsHash(),
	}

	return k.WriteConsensusState(ctx, ctx.BlockHeight(), cs)
}

// WriteConsensusState writes ConsensusState structure and corresponding height into the storage
func (k *Keeper) WriteConsensusState(ctx sdk.Context, height int64, cs tendermint.ConsensusState) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetConsensusStateKey(height)

	csBz, err := k.cdc.Marshal(&cs)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrJSONMarshal, err.Error())
	}

	store.Set(key, csBz)

	return nil
}

// Verify verifies that provided `values` are actually present on Neutron blockchain at `blockHeight`
func (k *Keeper) Verify(ctx sdk.Context, blockHeight int64, values []*icqtypes.StorageValue) error {
	store := ctx.KVStore(k.storeKey)

	// we need to use consensus state from the next height (N + 1), cause that consensus state contains .AppHash (Merkle Root) of the state for `blockHeight` (N)
	csBz := store.Get(types.GetConsensusStateKey(blockHeight + 1))
	if csBz == nil {
		return errors.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("consensus state for block %d not found", blockHeight))
	}

	var cs tendermint.ConsensusState
	if err := k.cdc.Unmarshal(csBz, &cs); err != nil {
		return errors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	if err := storageverification.VerifyStorageValues(values, cs.Root, ibccommitmenttypes.GetSDKSpecs(), nil); err != nil {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	return nil
}

// GetAllConsensusStates returns ALL consensus states that are present in the storage
// Pagination is not needed here because the method is used to export state to genesis
func (k *Keeper) GetAllConsensusStates(ctx sdk.Context) ([]*types.ConsensusState, error) {
	var (
		store  = prefix.NewStore(ctx.KVStore(k.storeKey), types.ConsensusStateKey)
		states []*types.ConsensusState
	)

	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		cs := tendermint.ConsensusState{}
		k.cdc.MustUnmarshal(iterator.Value(), &cs)
		height := int64(sdk.BigEndianToUint64(iterator.Key()))
		states = append(states, &types.ConsensusState{
			Height: height,
			Cs:     &cs,
		})
	}

	return states, nil
}
