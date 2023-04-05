package keeper

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	consumertypes "github.com/cosmos/interchain-security/x/ccv/consumer/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/neutron-org/neutron/x/feeburner/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace

		accountKeeper types.AccountKeeper
		bankKeeper    types.BankKeeper
	}
)

var KeyBurnedFees = []byte("BurnedFees")

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,

	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		paramstore:    ps,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
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
		totalBurnedNeutronsAmount.Coin = sdk.NewCoin(k.GetParams(ctx).NeutronDenom, sdk.NewInt(0))
	}

	return totalBurnedNeutronsAmount
}

// BurnAndDistribute is an important part of tokenomics. It does few things:
// 1. Burns NTRN fee coins distributed to consumertypes.ConsumerRedistributeName in ICS (https://github.com/cosmos/interchain-security/blob/v0.2.0/x/ccv/consumer/keeper/distribution.go#L17)
// 2. Updates total amount of burned NTRN coins
// 3. Sends non-NTRN fee tokens to treasury contract address
// Panics if no `consumertypes.ConsumerRedistributeName` module found OR could not burn NTRN tokens
func (k Keeper) BurnAndDistribute(ctx sdk.Context) {
	moduleAddr := k.accountKeeper.GetModuleAddress(consumertypes.ConsumerRedistributeName)
	if moduleAddr == nil {
		panic("ConsumerRedistributeName must have module address")
	}

	params := k.GetParams(ctx)
	balances := k.bankKeeper.GetAllBalances(ctx, moduleAddr)
	fundsForTreasury := make(sdk.Coins, 0, len(balances))

	for _, balance := range balances {
		if !balance.IsZero() {
			if balance.Denom == params.NeutronDenom {
				err := k.bankKeeper.BurnCoins(ctx, consumertypes.ConsumerRedistributeName, sdk.Coins{balance})
				if err != nil {
					panic(sdkerrors.Wrapf(err, "failed to burn NTRN tokens during fee processing"))
				}

				k.RecordBurnedFees(ctx, balance)
			} else {
				fundsForTreasury = append(fundsForTreasury, balance)
			}
		}
	}

	if len(fundsForTreasury) > 0 {
		addr, err := sdk.AccAddressFromBech32(params.TreasuryAddress)
		if err != nil {
			// there's no way we face this kind of situation in production, since it means the chain is misconfigured
			// still, in test environments it might be the case when the chain is started without treasury
			// in such case we just burn the tokens
			err := k.bankKeeper.BurnCoins(ctx, consumertypes.ConsumerRedistributeName, fundsForTreasury)
			if err != nil {
				panic(sdkerrors.Wrapf(err, "failed to burn tokens during fee processing"))
			}
		} else {
			err = k.bankKeeper.SendCoins(
				ctx,
				moduleAddr, addr,
				fundsForTreasury,
			)
			if err != nil {
				panic(sdkerrors.Wrapf(err, "failed sending funds to treasury"))
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
func (k Keeper) FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error {
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, authtypes.FeeCollectorName, amount)
}
