package v400

import (
	"context"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"encoding/base64"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v5/x/ccv/consumer/keeper"
	"github.com/neutron-org/neutron/v4/app/upgrades"
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

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

func createValidators(ctx sdk.Context, sk stakingkeeper.Keeper, consumerKeeper ccvconsumerkeeper.Keeper) error {
	// тут мы обнуляем всех ccv валидаторов
	for _, v := range consumerKeeper.GetAllCCValidator(ctx) {
		err := sk.SetLastValidatorPower(ctx, v.Address, 0)
		if err != nil {
			return fmt.Errorf("could not set last validator power for %s: %w", v.Address, err)
		}
	}

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
	pubkey, err := codectypes.NewAnyWithValue(pk)
	if err != nil {
		return err
	}

	srv := stakingkeeper.NewMsgServerImpl(&sk)
	_, err = srv.CreateValidator(ctx, &types.MsgCreateValidator{
		Description: types.Description{
			Moniker:         "sovereign",
			Identity:        "",
			Website:         "",
			SecurityContact: "",
			Details:         "",
		},
		Commission: types.CommissionRates{
			Rate:          math.LegacyMustNewDecFromStr("10.0"),
			MaxRate:       math.LegacyMustNewDecFromStr("10.0"),
			MaxChangeRate: math.LegacyMustNewDecFromStr("1.0"),
		},
		MinSelfDelegation: math.NewInt(1_000_000),
		DelegatorAddress:  "",
		ValidatorAddress:  "neutronvaloper18hl5c9xn5dze2g50uaw0l2mr02ew57zk5tccmr",
		Pubkey:            pubkey,
		// кто оплатит?
		Value: sdk.Coin{
			Denom:  "untrn",
			Amount: math.NewInt(1_000_000),
		},
	})
	if err != nil {
		return err
	}

	sk.SetLastTotalPower(ctx, math.NewInt(1))
	return nil
}
