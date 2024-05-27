package dynamicfees_test

//
//import (
//	"github.com/neutron-org/neutron/v4/x/feeburner"
//	"testing"
//
//	"github.com/neutron-org/neutron/v4/app/config"
//
//	"github.com/stretchr/testify/require"
//
//	"github.com/neutron-org/neutron/v4/testutil/common/nullify"
//	"github.com/neutron-org/neutron/v4/testutil/dynamicfees/keeper"
//	"github.com/neutron-org/neutron/v4/x/dynamicfees/types"
//)
//
//func TestGenesis(t *testing.T) {
//	_ = config.GetDefaultConfig()
//
//	genesisState := types.GenesisState{
//		Params: types.DefaultParams(),
//	}
//
//	k, ctx := keeper.FeeburnerKeeper(t)
//	feeburner.InitGenesis(ctx, *k, genesisState)
//	got := feeburner.ExportGenesis(ctx, *k)
//	require.NotNil(t, got)
//
//	nullify.Fill(&genesisState)
//	nullify.Fill(got)
//}
