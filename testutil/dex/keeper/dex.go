package keeper

import (
	"testing"

	cmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/neutron-org/neutron/x/dex/keeper"
	"github.com/neutron-org/neutron/x/dex/types"
	"github.com/stretchr/testify/require"
)

func DexKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := cmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	paramsSubspace := typesparams.NewSubspace(cdc,
		types.Amino,
		storeKey,
		memStoreKey,
		"DexParams",
	)
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		paramsSubspace,
		nil,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ctx
}

func AssertEventEmitted(t *testing.T, ctx sdk.Context, eventValue, message string) {
	allEvents := ctx.EventManager().Events()
	for _, event := range allEvents {
		for _, attr := range event.Attributes {
			if string(attr.Value) == eventValue {
				return
			}
		}
	}
	require.Fail(t, message)
}

func AssertNEventsEmitted(t *testing.T, ctx sdk.Context, eventValue string, nEvents int) {
	emissions := 0
	allEvents := ctx.EventManager().Events()
	for _, event := range allEvents {
		for _, attr := range event.Attributes {
			if string(attr.Value) == eventValue {
				emissions++
			}
		}
	}
	require.Equal(t, nEvents, emissions, "Expected %v events, got %v", nEvents, emissions)
}

func AssertEventNotEmitted(t *testing.T, ctx sdk.Context, eventValue, message string) {
	allEvents := ctx.EventManager().Events()
	if len(allEvents) != 0 {
		for _, attr := range allEvents[len(allEvents)-1].Attributes {
			if string(attr.Value) == eventValue {
				require.Fail(t, message)
			}
		}
	}
}
