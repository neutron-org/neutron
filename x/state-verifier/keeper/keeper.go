package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/header"
	"cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibccommitmenttypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	tendermint "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	ics23 "github.com/cosmos/ics23/go"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types2 "github.com/neutron-org/neutron/v5/x/interchainqueries/types"
	"github.com/neutron-org/neutron/v5/x/state-verifier/types"
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

func (k *Keeper) VerifyStateValues(ctx context.Context, request *types.QueryVefiryStateValuesRequest) (*types.QueryVerifyStateValuesResponse, error) {
	if err := k.Verify(sdk.UnwrapSDKContext(ctx), int64(request.Height), request.StorageValues); err != nil {
		return nil, err
	}

	return &types.QueryVerifyStateValuesResponse{Valid: true}, nil
}

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

func (k *Keeper) Verify(ctx sdk.Context, blockHeight int64, values []*types2.StorageValue) error {
	store := ctx.KVStore(k.storeKey)

	csBz := store.Get(types.GetConsensusStateKey(blockHeight + 1))
	if csBz == nil {
		return errors.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("consensus state for block %d not found", blockHeight))
	}

	var cs tendermint.ConsensusState
	if err := json.Unmarshal(csBz, &cs); err != nil {
		return errors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	for _, result := range values {
		proof, err := ibccommitmenttypes.ConvertProofs(result.Proof)
		if err != nil {
			return errors.Wrapf(sdkerrors.ErrInvalidType, "failed to convert crypto.ProofOps to MerkleProof: %v", err)
		}

		path := ibccommitmenttypes.NewMerklePath(result.StoragePrefix, string(result.Key))
		// identify what kind proofs (non-existence proof always has *ics23.CommitmentProof_Nonexist as the first item) we got
		// and call corresponding method to verify it
		switch proof.GetProofs()[0].GetProof().(type) {
		// we can get non-existence proof if someone queried some key which is not exists in the storage on remote chain
		case *ics23.CommitmentProof_Nonexist:
			if err := proof.VerifyNonMembership(ibccommitmenttypes.GetSDKSpecs(), cs.Root, path); err != nil {
				return errors.Wrapf(types2.ErrInvalidProof, "failed to verify proof: %v", err)
			}
			result.Value = nil
		case *ics23.CommitmentProof_Exist:
			if err := proof.VerifyMembership(ibccommitmenttypes.GetSDKSpecs(), cs.Root, path, result.Value); err != nil {
				return errors.Wrapf(types2.ErrInvalidProof, "failed to verify proof: %v", err)
			}
		default:
			return errors.Wrapf(types2.ErrInvalidProof, "unknown proof type %T", proof.GetProofs()[0].GetProof())
		}
	}

	return nil
}

func (k Keeper) GetAllConsensusStates(ctx sdk.Context) ([]*types.ConsensusState, error) {
	var (
		store  = prefix.NewStore(ctx.KVStore(k.storeKey), types.ConsensusStateKey)
		states []*types.ConsensusState
	)

	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		cs := tendermint.ConsensusState{}
		k.cdc.MustUnmarshal(iterator.Value(), &cs)
		height, err := strconv.ParseInt(string(iterator.Key()), 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to extract height from consensus state key")
		}

		states = append(states, &types.ConsensusState{
			Height: height,
			Cs:     &cs,
		})
	}

	return states, nil
}