package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/neutron-org/neutron/v6/x/feeburner/types"
)

type (
	Keeper struct {
		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey
		memKey   storetypes.StoreKey

		accountKeeper types.AccountKeeper
		bankKeeper    types.BankKeeper
		authority     string

		feeRedistributeName string
	}
)

var KeyBurnedFees = []byte("BurnedFees")

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	authority string,
	feeRedistributeName string,
) *Keeper {
	return &Keeper{
		cdc:                 cdc,
		storeKey:            storeKey,
		memKey:              memKey,
		accountKeeper:       accountKeeper,
		bankKeeper:          bankKeeper,
		authority:           authority,
		feeRedistributeName: feeRedistributeName,
	}
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

// RecordBurnedFees adds `amount` to the total amount of burned NTRN tokens
func (k Keeper) RecordBurnedFees(ctx sdk.Context, amount sdk.Coin) {
	store := ctx.KVStore(k.storeKey)

	totalBurnedNeutronsAmount := k.GetTotalBurnedNeutronsAmount(ctx)
	totalBurnedNeutronsAmount.Coin = totalBurnedNeutronsAmount.Coin.Add(amount)

	store.Set(KeyBurnedFees, k.cdc.MustMarshal(&totalBurnedNeutronsAmount))
}

// GetTotalBurnedNeutronsAmount gets the total burned amount of NTRN tokens
func (k Keeper) GetTotalBurnedNeutronsAmount(ctx sdk.Context) types.TotalBurnedNeutronsAmount {
	store := ctx.KVStore(k.storeKey)

	var totalBurnedNeutronsAmount types.TotalBurnedNeutronsAmount
	bzTotalBurnedNeutronsAmount := store.Get(KeyBurnedFees)
	if bzTotalBurnedNeutronsAmount != nil {
		k.cdc.MustUnmarshal(bzTotalBurnedNeutronsAmount, &totalBurnedNeutronsAmount)
	}

	if totalBurnedNeutronsAmount.Coin.Denom == "" {
		totalBurnedNeutronsAmount.Coin = sdk.NewCoin(k.GetParams(ctx).NeutronDenom, math.NewInt(0))
	}

	return totalBurnedNeutronsAmount
}

// SetTotalBurnedNeutronsAmount sets the total burned amount of NTRN tokens
func (k Keeper) SetTotalBurnedNeutronsAmount(ctx sdk.Context, totalBurnedNeutronsAmount types.TotalBurnedNeutronsAmount) {
	store := ctx.KVStore(k.storeKey)

	store.Set(KeyBurnedFees, k.cdc.MustMarshal(&totalBurnedNeutronsAmount))
}

// BurnAndDistribute is an important part of tokenomics. It does few things:
// 1. Burns NTRN fee coins distributed to fee redistribute module
// 2. Updates total amount of burned NTRN coins
// 3. Sends non-NTRN fee tokens to treasury contract address
// Panics if no fee redistribute module found OR could not burn NTRN tokens
func (k Keeper) BurnAndDistribute(ctx sdk.Context) {
	moduleAddr := k.accountKeeper.GetModuleAddress(k.feeRedistributeName)
	if moduleAddr == nil {
		panic("feeRedistributeName must have module address")
	}

	params := k.GetParams(ctx)
	balances := k.bankKeeper.GetAllBalances(ctx, moduleAddr)
	fundsForReserve := make(sdk.Coins, 0, len(balances))

	for _, balance := range balances {
		if !balance.IsZero() {
			if balance.Denom == params.NeutronDenom {
				err := k.bankKeeper.BurnCoins(ctx, k.feeRedistributeName, sdk.Coins{balance})
				if err != nil {
					panic(errors.Wrapf(err, "failed to burn NTRN tokens during fee processing"))
				}

				k.RecordBurnedFees(ctx, balance)
			} else {
				fundsForReserve = append(fundsForReserve, balance)
			}
		}
	}

	if len(fundsForReserve) > 0 {
		addr, err := sdk.AccAddressFromBech32(params.TreasuryAddress)
		if err != nil {
			// there's no way we face this kind of situation in production, since it means the chain is misconfigured
			// still, in test environments it might be the case when the chain is started without Reserve
			// in such case we just burn the tokens
			err := k.bankKeeper.BurnCoins(ctx, k.feeRedistributeName, fundsForReserve)
			if err != nil {
				panic(errors.Wrapf(err, "failed to burn tokens during fee processing"))
			}
		} else {
			err = k.bankKeeper.SendCoins(
				ctx,
				moduleAddr, addr,
				fundsForReserve,
			)
			if err != nil {
				panic(errors.Wrapf(err, "failed sending funds to Reserve"))
			}
		}
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// FundCommunityPool is method to satisfy DistributionKeeper interface for packet-forward-middleware Keeper.
// The original method sends coins to a community pool of a chain.
// The current method sends coins to a Fee Collector module which collects fee on consumer chain.
func (k Keeper) FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error {
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, authtypes.FeeCollectorName, amount)
}
