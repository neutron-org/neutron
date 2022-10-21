package sudo_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/neutron-org/neutron/internal/sudo"
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
	args := w.Called(ctx, contractAddress)
	buf, _ := args.Get(0).([]byte)
	return buf, args.Error(1)
}

type ContractManagerSuite struct {
	suite.Suite

	ctx             sdk.Context
	manager         sudo.ContractManager
	wasmKeeperMock  *WasmKeeperMock
	contractAddress sdk.AccAddress
}

func (suite *ContractManagerSuite) SetupTest() {
	key := sdk.NewKVStoreKey(suite.T().Name())
	suite.ctx = testutil.DefaultContext(key, sdk.NewTransientStoreKey("transient_"+suite.T().Name()))
	suite.manager = sudo.NewContractManager()
	suite.contractAddress, _ = sdk.AccAddressFromBech32("neutron1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqcd0mrx")

	suite.wasmKeeperMock = new(WasmKeeperMock)
	suite.manager.SetWasmKeeper(suite.wasmKeeperMock)
}

func (suite *ContractManagerSuite) TestHasContractInfoAbsentWasmKeeper() {
	manager := sudo.NewContractManager()
	defer func() {
		if r := recover(); r != nil {
			suite.Equal("wasmKeeper pointer is nil", r)
		}
	}()
	manager.HasContractInfo(suite.ctx, suite.contractAddress)

	suite.FailNow("The code did not panic")
}

func (suite *ContractManagerSuite) TestHasContractInfo() {
	suite.wasmKeeperMock.On("HasContractInfo", suite.ctx, suite.contractAddress).Return(true)
	result := suite.manager.HasContractInfo(suite.ctx, suite.contractAddress)

	suite.Require().True(result)
}

func (suite *ContractManagerSuite) TestSudoAbsentWasmKeeper() {
	manager := sudo.NewContractManager()
	defer func() {
		if r := recover(); r != nil {
			suite.Equal("wasmKeeper pointer is nil", r)
		}
	}()
	manager.Sudo(suite.ctx, suite.contractAddress, []byte{})

	suite.FailNow("The code did not panic")
}

func (suite *ContractManagerSuite) TestSudo() {
	suite.wasmKeeperMock.On("Sudo", suite.ctx, suite.contractAddress).Return([]byte{}, nil)
	result, err := suite.manager.Sudo(suite.ctx, suite.contractAddress, []byte{})

	suite.Require().Equal([]byte{}, result)
	suite.Require().Equal(nil, err)
}

func (suite *ContractManagerSuite) TestNewSudoHandler() {
	sudoHandler := suite.manager.NewSudoHandler("test module")

	suite.Require().IsType(sudo.Handler{}, sudoHandler)
}

func TestContractManagerSuite(t *testing.T) {
	suite.Run(t, new(ContractManagerSuite))
}
