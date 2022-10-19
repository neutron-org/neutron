package contractmanager

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ContractMethods interface {
	HasContractInfo(sdk.Context, sdk.AccAddress) bool
	Sudo(sdk.Context, sdk.AccAddress, []byte) ([]byte, error)
}

type ContractManager struct {
	wasmKeeper ContractMethods
}

var _ ContractMethods = (*ContractManager)(nil)

func NewContractManager() ContractManager {
	return ContractManager{}
}

func (cm *ContractManager) SetWasmKeeper(wasmKeeper ContractMethods) {
	cm.wasmKeeper = wasmKeeper
}

func (cm *ContractManager) GetWasmKeeper() ContractMethods {
	fmt.Println("SetWasmKeeper", cm.wasmKeeper)
	return cm.wasmKeeper

}

func (cm *ContractManager) HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool {
	if cm.wasmKeeper == nil {
		fmt.Println("HasContractInfo", cm.wasmKeeper)
		panic("wasmKeeper pointer is nil")
	}
	return cm.wasmKeeper.HasContractInfo(ctx, contractAddress)
}

func (cm *ContractManager) Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
	panic("Not implemented")
}
