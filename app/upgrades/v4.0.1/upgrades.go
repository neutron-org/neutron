package v400

import (
	"context"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"encoding/base64"
	"fmt"
	adminmoduletypes "github.com/cosmos/admin-module/v2/x/adminmodule/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v5/x/ccv/consumer/keeper"
	"github.com/neutron-org/neutron/v4/app/upgrades"
	"time"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.UpgradeKeepers,
	_ upgrades.StoreKeys,
	_ codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)

		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}
		err = createValidators(ctx, *keepers.StakingKeeper, *keepers.ConsumerKeeper)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

func createValidators(ctx sdk.Context, sk stakingkeeper.Keeper, consumerKeeper ccvconsumerkeeper.Keeper) error {
	srv := stakingkeeper.NewMsgServerImpl(&sk)
	micComm, err := math.LegacyNewDecFromStr("0.0")
	if err != nil {
		return err
	}
	params := types.Params{
		UnbondingTime:     21 * 24 * time.Hour,
		MaxValidators:     100,
		MaxEntries:        100,
		HistoricalEntries: 100,
		BondDenom:         "untrn",
		MinCommissionRate: micComm,
	}

	_, err = srv.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
		Params:    params,
	})
	if err != nil {
		return err
	}

	// тут мы добавляем всех ccv валидаторов в стейкинг модуль
	for _, v := range consumerKeeper.GetAllCCValidator(ctx) {
		fmt.Println(v.Address)

		add, err := bech32.ConvertAndEncode("neutronvaloper", v.GetAddress())
		if err != nil {
			return err
		}
		_, err = srv.CreateValidator(ctx, &types.MsgCreateValidator{
			Description: types.Description{
				Moniker:         "dd",
				Identity:        "",
				Website:         "",
				SecurityContact: "",
				Details:         "",
			},
			Commission: types.CommissionRates{
				Rate:          math.LegacyMustNewDecFromStr("0.1"),
				MaxRate:       math.LegacyMustNewDecFromStr("0.1"),
				MaxChangeRate: math.LegacyMustNewDecFromStr("0.1"),
			},
			MinSelfDelegation: math.NewInt(1_000_000),
			DelegatorAddress:  "",
			ValidatorAddress:  add,
			Pubkey:            v.GetPubkey(),
			// кто оплатит?
			Value: sdk.Coin{
				Denom:  "untrn",
				Amount: math.NewInt(1_000_000),
			},
		})
		if err != nil {
			return err
		}

	}
	_, b, _ := bech32.DecodeAndConvert("neutronvaloper18hl5c9xn5dze2g50uaw0l2mr02ew57zk5tccmr")
	fmt.Println(b)

	//pk1 := ed25519.GenPrivKey().PubKey()
	//require.NotNil(pk1)
	//
	//pubkey, err := codectypes.NewAnyWithValue(pk1)
	//require.NoError(err)
	pkraw, err := base64.StdEncoding.DecodeString("U5OsDjF61okt7TsPoM4NUokEACQ4KZCdGNnHYT8d36w=")
	if err != nil {
		return err
	}
	pk := ed25519.PubKey{Key: pkraw}
	pubkey, err := codectypes.NewAnyWithValue(&pk)
	if err != nil {
		return err
	}

	_, err = srv.CreateValidator(ctx, &types.MsgCreateValidator{
		Description: types.Description{
			Moniker:         "sovereign",
			Identity:        "",
			Website:         "",
			SecurityContact: "",
			Details:         "",
		},
		Commission: types.CommissionRates{
			Rate:          math.LegacyMustNewDecFromStr("0.1"),
			MaxRate:       math.LegacyMustNewDecFromStr("0.1"),
			MaxChangeRate: math.LegacyMustNewDecFromStr("0.1"),
		},
		MinSelfDelegation: math.NewInt(1_000_000),
		DelegatorAddress:  "",
		ValidatorAddress:  "neutronvaloper18hl5c9xn5dze2g50uaw0l2mr02ew57zk5tccmr",
		Pubkey:            pubkey,
		// кто оплатит?
		Value: sdk.Coin{
			Denom:  "untrn",
			Amount: math.NewInt(100_000_000),
		},
	})
	if err != nil {
		return err
	}

	sk.SetLastTotalPower(ctx, math.NewInt(1))
	return nil
}
