package keeper_test

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/testutil"
	dextypes "github.com/neutron-org/neutron/x/dex/types"
	. "github.com/neutron-org/neutron/x/incentives/keeper"
	"github.com/neutron-org/neutron/x/incentives/types"
)

func TestCombineKeys(t *testing.T) {
	// create three keys, each different byte arrays
	key1 := []byte{0x11}
	key2 := []byte{0x12}
	key3 := []byte{0x13}

	// combine the three keys into a single key
	key := types.CombineKeys(key1, key2, key3)

	// three keys plus two separators is equal to a length of 5
	require.Len(t, key, 3+2)

	// ensure the newly created key is made up of the three previous keys (and the two key index separators)
	require.Equal(t, key[0], key1[0])
	require.Equal(t, key[1], types.KeyIndexSeparator[0])
	require.Equal(t, key[2], key2[0])
	require.Equal(t, key[3], types.KeyIndexSeparator[0])
	require.Equal(t, key[4], key3[0])
}

func TestFindIndex(t *testing.T) {
	// create an array of 5 IDs
	IDs := []uint64{1, 2, 3, 4, 5}

	// use the FindIndex function to find the index of the respective IDs
	// if it doesn't exist, return -1
	require.Equal(t, FindIndex(IDs, 1), 0)
	require.Equal(t, FindIndex(IDs, 3), 2)
	require.Equal(t, FindIndex(IDs, 5), 4)
	require.Equal(t, FindIndex(IDs, 6), -1)
}

func TestRemoveValue(t *testing.T) {
	// create an array of 5 IDs
	IDs := []uint64{1, 2, 3, 4, 5}

	// remove an ID
	// ensure if ID exists, the length is reduced by one and the index of the removed ID is returned
	IDs, index1 := RemoveValue(IDs, 5)
	require.Len(t, IDs, 4)
	require.Equal(t, index1, 4)
	IDs, index2 := RemoveValue(IDs, 3)
	require.Len(t, IDs, 3)
	require.Equal(t, index2, 2)
	IDs, index3 := RemoveValue(IDs, 1)
	require.Len(t, IDs, 2)
	require.Equal(t, index3, 0)
	IDs, index4 := RemoveValue(IDs, 6)
	require.Len(t, IDs, 2)
	require.Equal(t, index4, -1)
}

func TestStakeRefKeys(t *testing.T) {
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	app := testutil.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	pool1, err := app.DexKeeper.InitPool(ctx, &dextypes.PairID{Token0: "TokenA", Token1: "TokenB"}, 0, 1)
	require.NoError(t, err)

	denom1 := pool1.GetPoolDenom()

	pool2, err := app.DexKeeper.InitPool(ctx, &dextypes.PairID{Token0: "TokenA", Token1: "TokenC"}, 0, 1)
	require.NoError(t, err)

	denom2 := pool2.GetPoolDenom()

	// empty address and 1 coin
	stake1 := types.NewStake(
		1,
		sdk.AccAddress{},
		sdk.Coins{sdk.NewInt64Coin(denom1, 10)},
		time.Now(),
		10,
	)
	_, err = app.IncentivesKeeper.GetStakeRefKeys(ctx, stake1)
	require.Error(t, err)

	// empty address and 2 coins
	stake2 := types.NewStake(
		1,
		sdk.AccAddress{},
		sdk.Coins{sdk.NewInt64Coin(denom1, 10), sdk.NewInt64Coin(denom2, 1)},
		time.Now(),
		10,
	)
	_, err = app.IncentivesKeeper.GetStakeRefKeys(ctx, stake2)
	require.Error(t, err)

	// not empty address and 1 coin
	stake3 := types.NewStake(1, addr1, sdk.Coins{sdk.NewInt64Coin(denom1, 10)}, time.Now(), 10)
	keys3, err := app.IncentivesKeeper.GetStakeRefKeys(ctx, stake3)
	require.NoError(t, err)
	require.Len(t, keys3, 6)

	// not empty address and empty coin
	stake4 := types.NewStake(1, addr1, sdk.Coins{sdk.NewInt64Coin(denom1, 10)}, time.Now(), 10)
	keys4, err := app.IncentivesKeeper.GetStakeRefKeys(ctx, stake4)
	require.NoError(t, err)
	require.Len(t, keys4, 6)

	// not empty address and 2 coins
	stake5 := types.NewStake(
		1,
		addr1,
		sdk.Coins{sdk.NewInt64Coin(denom1, 10), sdk.NewInt64Coin(denom2, 1)},
		time.Now(),
		10,
	)
	keys5, err := app.IncentivesKeeper.GetStakeRefKeys(ctx, stake5)
	require.NoError(t, err)
	require.Len(t, keys5, 10)
}
