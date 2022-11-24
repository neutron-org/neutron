package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/x/feerefunder/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
)

func TestKeeperCheckFees(t *testing.T) {
	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := tmdb.NewMemDB()
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
		"FeeParams",
	)
	k := NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		paramsSubspace,
		nil,
		nil,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	k.SetParams(ctx, types.Params{
		MinFee: types.Fee{
			RecvFee:    nil,
			AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(100)), sdk.NewCoin("denom2", sdk.NewInt(100))),
			TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(100)), sdk.NewCoin("denom2", sdk.NewInt(100))),
		},
	})

	for _, tc := range []struct {
		desc  string
		fees  *types.Fee
		valid bool
	}{
		{
			desc: "single proper denom but insufficient",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(1))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(1))),
			},
			valid: false,
		},
		{
			desc: "single denom sufficient amount",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(101))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(101))),
			},
			valid: true,
		},
		{
			desc: "multiple denoms, both are proper, only one enough",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(101)), sdk.NewCoin("denom2", sdk.NewInt(1))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(101)), sdk.NewCoin("denom2", sdk.NewInt(1))),
			},
			valid: true,
		},
		{
			desc: "no proper denom",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom3", sdk.NewInt(1))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom3", sdk.NewInt(1))),
			},
			valid: false,
		},
		{
			desc: "proper denom plus random one",
			fees: &types.Fee{
				RecvFee:    nil,
				AckFee:     sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(101)), sdk.NewCoin("denom3", sdk.NewInt(1))),
				TimeoutFee: sdk.NewCoins(sdk.NewCoin("denom1", sdk.NewInt(101)), sdk.NewCoin("denom3", sdk.NewInt(1))),
			},
			valid: true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := k.checkFees(ctx, *tc.fees)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
