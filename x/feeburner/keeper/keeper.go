package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/neutron-org/neutron/x/feeburner/types"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace

		bankKeeper types.BankKeeper
	}
)

var KeyBurnedFees = []byte("BurnedFees")

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,

	bankKeeper types.BankKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
		bankKeeper: bankKeeper,
	}
}

func (k Keeper) RecordBurnedFees(ctx sdk.Context, amount sdk.Coin) {
	store := ctx.KVStore(k.storeKey)

	totalBurnedNeutronsAmount := k.GetTotalBurnedNeutronsAmount(ctx)
	totalBurnedNeutronsAmount.Coins = totalBurnedNeutronsAmount.Coins.Add(amount)
	store.Set(KeyBurnedFees, k.cdc.MustMarshal(&totalBurnedNeutronsAmount))
}

func (k Keeper) GetTotalBurnedNeutronsAmount(ctx sdk.Context) types.TotalBurnedNeutronsAmount {
	store := ctx.KVStore(k.storeKey)

	var totalBurnedNeutronsAmount types.TotalBurnedNeutronsAmount
	bzTotalBurnedNeutronsAmount := store.Get(KeyBurnedFees)
	if bzTotalBurnedNeutronsAmount != nil {
		k.cdc.MustUnmarshal(bzTotalBurnedNeutronsAmount, &totalBurnedNeutronsAmount)
	}

	return totalBurnedNeutronsAmount
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
