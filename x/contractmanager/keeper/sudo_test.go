package keeper_test

import (
	"encoding/json"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/app"
	"github.com/neutron-org/neutron/testutil"
	keepertest "github.com/neutron-org/neutron/testutil/contractmanager/keeper"
	mock_types "github.com/neutron-org/neutron/testutil/mocks/contractmanager/types"
	"github.com/neutron-org/neutron/x/contractmanager/types"
)

func init() {
	app.GetDefaultConfig()
}

func TestSudoHasAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	wk := mock_types.NewMockWasmKeeper(ctrl)

	k, ctx := keepertest.ContractManagerKeeper(t, wk)
	address := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)

	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(true)
	has := k.HasContractInfo(ctx, address)
	require.Equal(t, true, has)

	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(false)
	has = k.HasContractInfo(ctx, address)
	require.Equal(t, false, has)
}

func TestSudoResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	wk := mock_types.NewMockWasmKeeper(ctrl)

	k, ctx := keepertest.ContractManagerKeeper(t, wk)
	address := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)

	sudoErrorMsg := types.MessageResponse{}
	p := channeltypes.Packet{}
	a := channeltypes.Acknowledgement{Response: &channeltypes.Acknowledgement_Result{Result: []byte("data")}}
	sudoErrorMsg.Response.Data = a.GetResult()
	sudoErrorMsg.Response.Request = p
	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoErrorMsg)).Return([]byte("success"), nil)
	resp, err := k.SudoResponse(ctx, address, sudoErrorMsg.Response.Request, a)
	require.NoError(t, err)
	require.Equal(t, []byte("success"), resp)

	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoErrorMsg)).Return(nil, fmt.Errorf("internal contract error"))
	resp, err = k.SudoResponse(ctx, address, sudoErrorMsg.Response.Request, a)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "internal contract error")
}

func TestSudoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	wk := mock_types.NewMockWasmKeeper(ctrl)

	k, ctx := keepertest.ContractManagerKeeper(t, wk)
	address := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)

	sudoErrorMsg := types.MessageError{}
	p := channeltypes.Packet{}
	a := channeltypes.Acknowledgement{Response: &channeltypes.Acknowledgement_Error{
		Error: "details",
	}}
	sudoErrorMsg.Error.Details = a.GetError()
	sudoErrorMsg.Error.Request = p
	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoErrorMsg)).Return([]byte("success"), nil)
	resp, err := k.SudoError(ctx, address, sudoErrorMsg.Error.Request, a)
	require.NoError(t, err)
	require.Equal(t, []byte("success"), resp)

	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoErrorMsg)).Return(nil, fmt.Errorf("internal contract error"))
	resp, err = k.SudoError(ctx, address, sudoErrorMsg.Error.Request, a)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "internal contract error")
}

func TestSudoTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	wk := mock_types.NewMockWasmKeeper(ctrl)

	k, ctx := keepertest.ContractManagerKeeper(t, wk)
	address := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)

	sudoTimeoutMsg := types.MessageTimeout{}
	p := channeltypes.Packet{}
	sudoTimeoutMsg.Timeout.Request = p
	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoTimeoutMsg)).Return([]byte("success"), nil)
	resp, err := k.SudoTimeout(ctx, address, sudoTimeoutMsg.Timeout.Request)
	require.NoError(t, err)
	require.Equal(t, []byte("success"), resp)

	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoTimeoutMsg)).Return(nil, fmt.Errorf("internal contract error"))
	resp, err = k.SudoTimeout(ctx, address, sudoTimeoutMsg.Timeout.Request)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "internal contract error")
}

func TestSudoOnChanOpen(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	wk := mock_types.NewMockWasmKeeper(ctrl)

	k, ctx := keepertest.ContractManagerKeeper(t, wk)
	address := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)

	sudoOpenAckMsg := types.MessageOnChanOpenAck{}
	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoOpenAckMsg)).Return([]byte("success"), nil)
	resp, err := k.SudoOnChanOpenAck(ctx, address, sudoOpenAckMsg.OpenAck)
	require.NoError(t, err)
	require.Equal(t, []byte("success"), resp)

	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoOpenAckMsg)).Return(nil, fmt.Errorf("internal contract error"))
	resp, err = k.SudoOnChanOpenAck(ctx, address, sudoOpenAckMsg.OpenAck)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "internal contract error")
}

func TestSudoTxQueryResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	wk := mock_types.NewMockWasmKeeper(ctrl)

	k, ctx := keepertest.ContractManagerKeeper(t, wk)
	address := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)

	sudoTxQueryResultMsg := types.MessageTxQueryResult{}
	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoTxQueryResultMsg)).Return([]byte("success"), nil)
	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(true)
	resp, err := k.SudoTxQueryResult(ctx,
		address,
		sudoTxQueryResultMsg.TxQueryResult.QueryID,
		sudoTxQueryResultMsg.TxQueryResult.Height,
		sudoTxQueryResultMsg.TxQueryResult.Data,
	)
	require.NoError(t, err)
	require.Equal(t, []byte("success"), resp)

	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoTxQueryResultMsg)).Return(nil, fmt.Errorf("internal contract error"))
	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(true)
	resp, err = k.SudoTxQueryResult(ctx,
		address,
		sudoTxQueryResultMsg.TxQueryResult.QueryID,
		sudoTxQueryResultMsg.TxQueryResult.Height,
		sudoTxQueryResultMsg.TxQueryResult.Data,
	)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "internal contract error")

	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(false)
	resp, err = k.SudoTxQueryResult(ctx,
		address,
		sudoTxQueryResultMsg.TxQueryResult.QueryID,
		sudoTxQueryResultMsg.TxQueryResult.Height,
		sudoTxQueryResultMsg.TxQueryResult.Data,
	)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "is not a contract address")
}

func TestSudoKvQueryResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	wk := mock_types.NewMockWasmKeeper(ctrl)

	k, ctx := keepertest.ContractManagerKeeper(t, wk)
	address := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)

	sudoTxQueryResultMsg := types.MessageKVQueryResult{}
	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoTxQueryResultMsg)).Return([]byte("success"), nil)
	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(true)
	resp, err := k.SudoKVQueryResult(ctx,
		address,
		sudoTxQueryResultMsg.KVQueryResult.QueryID,
	)
	require.NoError(t, err)
	require.Equal(t, []byte("success"), resp)

	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoTxQueryResultMsg)).Return(nil, fmt.Errorf("internal contract error"))
	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(true)
	resp, err = k.SudoKVQueryResult(ctx,
		address,
		sudoTxQueryResultMsg.KVQueryResult.QueryID,
	)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "internal contract error")

	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(false)
	resp, err = k.SudoKVQueryResult(ctx,
		address,
		sudoTxQueryResultMsg.KVQueryResult.QueryID,
	)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "is not a contract address")
}

func mustJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
