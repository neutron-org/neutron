package contractmanager_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/internal/contractmanager"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type WasmKeeperMock struct {
	mock.Mock
}

func (w *WasmKeeperMock) HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool {
	args := w.Called(ctx, contractAddress)
	return args.Bool(0)
}

func (w *WasmKeeperMock) Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
	fmt.Println("Sudo")
	return msg, nil
}

type ContractManagerSuite struct {
	suite.Suite

	ctx     sdk.Context
	manager contractmanager.ContractManager
}

func (suite *ContractManagerSuite) SetupTest() {
	key := sdk.NewKVStoreKey(suite.T().Name())
	suite.ctx = testutil.DefaultContext(key, sdk.NewTransientStoreKey("transient_"+suite.T().Name()))
	suite.manager = contractmanager.NewContractManager()

	wasmKeeperMock := new(WasmKeeperMock)
	suite.manager.SetWasmKeeper(wasmKeeperMock)
}

func (suite *ContractManagerSuite) TestAbsentWasmKeeper() {
	manager := contractmanager.NewContractManager()
	contract, _ := sdk.AccAddressFromBech32("neutron1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqcd0mrx")
	defer func() {
		if r := recover(); r != nil {
			suite.Equal("wasmKeeper pointer is nil", r)
		}
	}()
	manager.HasContractInfo(suite.ctx, contract)

	suite.FailNow("The code did not panic")
}

func (suite *ContractManagerSuite) TestHasContractInfo() {
	contract, _ := sdk.AccAddressFromBech32("neutron1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqcd0mrx")
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in ", r)
		}
	}()
	suite.manager.HasContractInfo(suite.ctx, contract)

	suite.FailNow("The code did not panic")
}

func TestContractManagerSuite(t *testing.T) {
	suite.Run(t, new(ContractManagerSuite))
}
