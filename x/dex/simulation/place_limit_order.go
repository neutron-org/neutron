package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/neutron-org/neutron/v6/x/dex/keeper"
	"github.com/neutron-org/neutron/v6/x/dex/types"
)

func SimulateMsgPlaceLimitOrder(
	_ types.BankKeeper,
	_ keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, _ *baseapp.BaseApp, _ sdk.Context, accs []simtypes.Account, _ string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgPlaceLimitOrder{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the PlaceLimitOrder simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "PlaceLimitOrder simulation not implemented"), nil, nil
	}
}
