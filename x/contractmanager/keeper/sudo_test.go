package keeper_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/neutron-org/neutron/v6/app/config"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v6/testutil"
	keepertest "github.com/neutron-org/neutron/v6/testutil/contractmanager/keeper"
	mock_types "github.com/neutron-org/neutron/v6/testutil/mocks/contractmanager/types"
	"github.com/neutron-org/neutron/v6/x/contractmanager/types"
)

func init() {
	config.GetDefaultConfig()
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
