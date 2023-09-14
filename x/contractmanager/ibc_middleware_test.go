package contractmanager

//
//import (
//	tmdb "github.com/cometbft/cometbft-db"
//	"github.com/cometbft/cometbft/libs/log"
//	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
//	"github.com/cosmos/cosmos-sdk/store"
//	storetypes "github.com/cosmos/cosmos-sdk/store/types"
//	sdk "github.com/cosmos/cosmos-sdk/types"
//	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
//	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
//	"github.com/golang/mock/gomock"
//	mock_types "github.com/neutron-org/neutron/testutil/mocks/interchaintxs/types"
//	"github.com/neutron-org/neutron/x/contractmanager/types"
//	"github.com/stretchr/testify/require"
//	"testing"
//)
//
//var (
//	ShouldNotBeWrittenKey = []byte("shouldnotkey")
//	ShouldNotBeWritten    = []byte("should not be written")
//	ShouldBeWritten       = []byte("should be written")
//	TestOwnerAddress      = "neutron17dtl0mjt3t77kpuhg2edqzjpszulwhgzcdvagh"
//)
//
//func ShouldBeWrittenKey(suffix string) []byte {
//	return append([]byte("shouldkey"), []byte(suffix)...)
//}
//
//func NewSudoLimitMiddleware(t testing.TB, cm types.SudoWrapper) (SudoLimitWrapper, sdk.Context, *storetypes.KVStoreKey) {
//	storeKey := sdk.NewKVStoreKey(types.StoreKey)
//
//	db := tmdb.NewMemDB()
//	stateStore := store.NewCommitMultiStore(db)
//	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
//	require.NoError(t, stateStore.LoadLatestVersion())
//
//	k := SudoLimitWrapper{SudoWrapper: cm}
//
//	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())
//
//	return k, ctx, storeKey
//}
//
//func TestSudo(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//	cmKeeper := mock_types.NewMockContractManagerKeeper(ctrl)
//	middleware, infCtx, storeKey := NewSudoLimitMiddleware(t, cmKeeper)
//	st := infCtx.KVStore(storeKey)
//
//	p := channeltypes.Packet{
//		Sequence:      100,
//		SourcePort:    icatypes.ControllerPortPrefix + TestOwnerAddress + ".ica0",
//		SourceChannel: "channel-0",
//	}
//	contractAddress := sdk.AccAddress{}
//	errACK := channeltypes.Acknowledgement{
//		Response: &channeltypes.Acknowledgement_Error{
//			Error: "error",
//		},
//	}
//	//errAckData, err := channeltypes.SubModuleCdc.MarshalJSON(&errACK)
//	//require.NoError(t, err)
//	//resACK := channeltypes.Acknowledgement{
//	//	Response: &channeltypes.Acknowledgement_Result{Result: []byte("Result")},
//	//}
//	//resAckData, err := channeltypes.SubModuleCdc.MarshalJSON(&resACK)
//	//require.NoError(t, err)
//
//	//  success during SudoError
//	ctx := infCtx.WithGasMeter(sdk.NewGasMeter(1_000_000_000_000))
//	cmKeeper.EXPECT().SudoError(gomock.AssignableToTypeOf(ctx), contractAddress, p, errACK).Do(func(cachedCtx sdk.Context, senderAddress sdk.AccAddress, request channeltypes.Packet, errAck channeltypes.Acknowledgement) {
//		st := cachedCtx.KVStore(storeKey)
//		st.Set(ShouldBeWrittenKey("sudoerror"), ShouldBeWritten)
//	}).Return(nil, nil)
//	cmKeeper.EXPECT().GetParams(ctx).Return(types.Params{SudoCallGasLimit: 6000})
//	err := middleware.Sudo(ctx, contractAddress, p, &errACK)
//	require.NoError(t, err)
//	require.Equal(t, ShouldBeWritten, st.Get(ShouldBeWrittenKey("sudoerror")))
//	require.Equal(t, uint64(3050), ctx.GasMeter().GasConsumed())
//}
