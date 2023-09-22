package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/neutron-org/neutron/x/incentives/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Keeper provides a way to manage incentives module storage.
type Keeper struct {
	storeKey    storetypes.StoreKey
	paramSpace  paramtypes.Subspace
	hooks       types.IncentiveHooks
	ak          types.AccountKeeper
	bk          types.BankKeeper
	ek          types.EpochKeeper
	dk          types.DexKeeper
	distributor Distributor
	authority   string
}

// NewKeeper returns a new instance of the incentive module keeper struct.
func NewKeeper(
	storeKey storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	ek types.EpochKeeper,
	dk types.DexKeeper,
	authority string,
) *Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	keeper := &Keeper{
		storeKey:   storeKey,
		paramSpace: paramSpace,
		ak:         ak,
		bk:         bk,
		ek:         ek,
		dk:         dk,
		authority:  authority,
	}
	keeper.distributor = NewDistributor(keeper)
	return keeper
}

// SetHooks sets the incentives hooks.
func (k *Keeper) SetHooks(ih types.IncentiveHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set incentive hooks twice")
	}

	k.hooks = ih

	return k
}

// Logger returns a logger instance for the incentives module.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetModuleBalance returns full balance of the module.
func (k Keeper) GetModuleBalance(ctx sdk.Context) sdk.Coins {
	acc := k.ak.GetModuleAccount(ctx, types.ModuleName)
	return k.bk.GetAllBalances(ctx, acc.GetAddress())
}

// GetModuleStakedCoins Returns staked balance of the module.
func (k Keeper) GetModuleStakedCoins(ctx sdk.Context) sdk.Coins {
	// all not unstaking + not finished unstaking
	return k.GetStakes(ctx).GetCoins()
}
