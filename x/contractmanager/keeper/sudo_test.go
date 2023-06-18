package keeper_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
	sudoErrorMsg.Response.Data = []byte("data")
	sudoErrorMsg.Response.Request = p
	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoErrorMsg)).Return([]byte("success"), nil)
	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(true)
	resp, err := k.SudoResponse(ctx, address, sudoErrorMsg.Response.Request, sudoErrorMsg.Response.Data)
	require.NoError(t, err)
	require.Equal(t, []byte("success"), resp)

	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoErrorMsg)).Return(nil, fmt.Errorf("internal contract error"))
	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(true)
	resp, err = k.SudoResponse(ctx, address, sudoErrorMsg.Response.Request, sudoErrorMsg.Response.Data)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "internal contract error")

	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(false)
	resp, err = k.SudoResponse(ctx, address, channeltypes.Packet{}, nil)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "is not a contract address and not the Transfer module")

	sudoResponseTransport := types.MessageResponse{}
	p = channeltypes.Packet{SourcePort: types.TransferPort}
	sudoResponseTransport.Response.Data = []byte("data")
	sudoResponseTransport.Response.Request = p

	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(false)
	_, err = k.SudoResponse(ctx, address, sudoResponseTransport.Response.Request, sudoResponseTransport.Response.Data)
	require.Nil(t, err)
	require.NoError(t, err)
}

func TestSudoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	wk := mock_types.NewMockWasmKeeper(ctrl)

	k, ctx := keepertest.ContractManagerKeeper(t, wk)
	address := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)

	sudoErrorMsg := types.MessageError{}
	p := channeltypes.Packet{}
	sudoErrorMsg.Error.Details = "details"
	sudoErrorMsg.Error.Request = p
	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoErrorMsg)).Return([]byte("success"), nil)
	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(true)
	resp, err := k.SudoError(ctx, address, sudoErrorMsg.Error.Request, sudoErrorMsg.Error.Details)
	require.NoError(t, err)
	require.Equal(t, []byte("success"), resp)

	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoErrorMsg)).Return(nil, fmt.Errorf("internal contract error"))
	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(true)
	resp, err = k.SudoError(ctx, address, sudoErrorMsg.Error.Request, sudoErrorMsg.Error.Details)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "internal contract error")

	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(false)
	resp, err = k.SudoError(ctx, address, channeltypes.Packet{}, "")
	require.Nil(t, resp)
	require.ErrorContains(t, err, "is not a contract address and not the Transfer module")

	sudoErrorTransport := types.MessageError{}
	p = channeltypes.Packet{SourcePort: types.TransferPort}
	sudoErrorTransport.Error.Details = "details"
	sudoErrorTransport.Error.Request = p

	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(false)
	resp, err = k.SudoError(ctx, address, sudoErrorTransport.Error.Request, sudoErrorTransport.Error.Details)
	require.Nil(t, resp)
	require.NoError(t, err)
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
	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(true)
	resp, err := k.SudoTimeout(ctx, address, sudoTimeoutMsg.Timeout.Request)
	require.NoError(t, err)
	require.Equal(t, []byte("success"), resp)

	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoTimeoutMsg)).Return(nil, fmt.Errorf("internal contract error"))
	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(true)
	resp, err = k.SudoTimeout(ctx, address, sudoTimeoutMsg.Timeout.Request)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "internal contract error")

	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(false)
	resp, err = k.SudoTimeout(ctx, address, channeltypes.Packet{})
	require.Nil(t, resp)
	require.ErrorContains(t, err, "is not a contract address and not the Transfer module")

	sudoTimeoutTransport := types.MessageTimeout{}
	p = channeltypes.Packet{SourcePort: types.TransferPort}
	sudoTimeoutTransport.Timeout.Request = p

	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(false)
	resp, err = k.SudoTimeout(ctx, address, sudoTimeoutTransport.Timeout.Request)
	require.Nil(t, resp)
	require.NoError(t, err)
}

func TestSudoOnChanOpen(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	wk := mock_types.NewMockWasmKeeper(ctrl)

	k, ctx := keepertest.ContractManagerKeeper(t, wk)
	address := sdk.MustAccAddressFromBech32(testutil.TestOwnerAddress)

	sudoOpenAckMsg := types.MessageOnChanOpenAck{}
	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoOpenAckMsg)).Return([]byte("success"), nil)
	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(true)
	resp, err := k.SudoOnChanOpenAck(ctx, address, sudoOpenAckMsg.OpenAck)
	require.NoError(t, err)
	require.Equal(t, []byte("success"), resp)

	wk.EXPECT().Sudo(gomock.Any(), address, mustJSON(sudoOpenAckMsg)).Return(nil, fmt.Errorf("internal contract error"))
	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(true)
	resp, err = k.SudoOnChanOpenAck(ctx, address, sudoOpenAckMsg.OpenAck)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "internal contract error")

	wk.EXPECT().HasContractInfo(gomock.Any(), address).Return(false)
	resp, err = k.SudoOnChanOpenAck(ctx, address, sudoOpenAckMsg.OpenAck)
	require.Nil(t, resp)
	require.ErrorContains(t, err, "is not a contract address")
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
