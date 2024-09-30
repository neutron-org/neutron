package v400

import (
	"context"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"encoding/base64"
	"encoding/json"
	"fmt"
	adminmoduletypes "github.com/cosmos/admin-module/v2/x/adminmodule/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v5/x/ccv/consumer/keeper"
	"github.com/neutron-org/neutron/v5/app/upgrades"
	"os"
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
		err = createValidators(ctx, *keepers.StakingKeeper, *keepers.ConsumerKeeper, keepers.BankKeeper, keepers.AccountKeeper)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}

type PK struct {
	Address string `json:"address"`
	PubKey  struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"pub_key"`
	PrivKey struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"priv_key"`
}

func NewSovereignVal(ctx context.Context, id int, bk bankkeeper.Keeper, ak authkeeper.AccountKeeperI) types.MsgCreateValidator {
	pkpath := fmt.Sprintf("/home/swelf/src/lido/neutron/data/test-1/node-%d/config/priv_validator_key.json", id)
	pkdata, err := os.ReadFile(pkpath)
	if err != nil {
		panic(err)
	}
	pkraw := PK{}
	err = json.Unmarshal(pkdata, &pkraw)
	if err != nil {
		panic(err)
	}
	pkbytes, err := base64.StdEncoding.DecodeString(pkraw.PubKey.Value)
	if err != nil {
		panic(err)
	}
	pk := ed25519.PubKey{Key: pkbytes}
	pubkey, err := codectypes.NewAnyWithValue(&pk)
	if err != nil {
		panic(err)
	}

	// и конечно генерация сломала мне консенсус
	privBank := ed25519.GenPrivKey()
	privBank.PubKey().Address()

	err = bk.MintCoins(ctx, "dex", sdk.NewCoins(sdk.Coin{
		Denom:  "untrn",
		Amount: math.NewInt(1_000_000),
	}))
	if err != nil {
		panic(err)
	}

	err = bk.SendCoinsFromModuleToAccount(ctx, "dex", privBank.PubKey().Address().Bytes(), sdk.NewCoins(sdk.Coin{
		Denom:  "untrn",
		Amount: math.NewInt(1_000_000),
	}))
	if err != nil {
		panic(err)
	}

	add, err := bech32.ConvertAndEncode("neutronvaloper", privBank.PubKey().Address().Bytes())
	if err != nil {
		panic(err)
	}

	return types.MsgCreateValidator{
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
		// WARN: у оператора должно быть достаточно денег для selfbond
		ValidatorAddress: add,
		Pubkey:           pubkey,
		// кто оплатит?
		Value: sdk.Coin{
			Denom:  "untrn",
			Amount: math.NewInt(1_000_000),
		},
	}
}

func createValidators(ctx sdk.Context, sk stakingkeeper.Keeper, consumerKeeper ccvconsumerkeeper.Keeper, bk bankkeeper.Keeper, ak authkeeper.AccountKeeperI) error {
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
		//fmt.Println(v.Address)

		//bankAddress, err := bech32.ConvertAndEncode("neutron", v.GetAddress())
		//if err != nil {
		//	return err
		//}
		err = bk.MintCoins(ctx, "dex", sdk.NewCoins(sdk.Coin{
			Denom:  "untrn",
			Amount: math.NewInt(1_000_000),
		}))
		if err != nil {
			return err
		}

		err = bk.SendCoinsFromModuleToAccount(ctx, "dex", v.GetAddress(), sdk.NewCoins(sdk.Coin{
			Denom:  "untrn",
			Amount: math.NewInt(1_000_000),
		}))
		if err != nil {
			return err
		}

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
			// WARN: у оператора должно быть достаточно денег для selfbond
			ValidatorAddress: add,
			//ValidatorAddress: "neutronvaloper1m9l358xunhhwds0568za49mzhvuxx9uxamysqw",
			Pubkey: v.GetPubkey(),
			Value: sdk.Coin{
				Denom:  "untrn",
				Amount: math.NewInt(1_000_000),
			},
		})
		if err != nil {
			return err
		}

		err = sk.SetLastValidatorPower(ctx, v.GetAddress(), 1)
		if err != nil {
			return err
		}

		savedVal, err := sk.GetValidator(ctx, v.GetAddress())
		if err != nil {
			return err
		}
		_, err = bondValidator(ctx, sk, savedVal)
		if err != nil {
			return err
		}

	}
	//_, b, _ := bech32.DecodeAndConvert("neutronvaloper18hl5c9xn5dze2g50uaw0l2mr02ew57zk5tccmr")
	//fmt.Println(b)

	//pkraw, err := base64.StdEncoding.DecodeString("U5OsDjF61okt7TsPoM4NUokEACQ4KZCdGNnHYT8d36w=")
	//if err != nil {
	//	return err
	//}
	//pk := ed25519.PubKey{Key: pkraw}
	//pubkey, err := codectypes.NewAnyWithValue(&pk)
	//if err != nil {
	//	return err
	//}

	for i := 13; i <= 14; i++ {
		msgCreate := NewSovereignVal(ctx, i, bk, ak)
		_, err = srv.CreateValidator(ctx, &msgCreate)
		if err != nil {
			return err
		}
	}

	sk.SetLastTotalPower(ctx, math.NewInt(1))
	return nil
}

func bondValidator(ctx context.Context, k stakingkeeper.Keeper, validator types.Validator) (types.Validator, error) {
	// delete the validator by power index, as the key will change
	if err := k.DeleteValidatorByPowerIndex(ctx, validator); err != nil {
		return validator, err
	}

	validator = validator.UpdateStatus(types.Bonded)

	// save the now bonded validator record to the two referenced stores
	if err := k.SetValidator(ctx, validator); err != nil {
		return validator, err
	}

	if err := k.SetValidatorByPowerIndex(ctx, validator); err != nil {
		return validator, err
	}

	// delete from queue if present
	if err := k.DeleteValidatorQueue(ctx, validator); err != nil {
		return validator, err
	}

	// trigger hook
	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return validator, err
	}
	codec := address.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix())
	str, err := codec.StringToBytes(validator.GetOperator())
	if err != nil {
		return validator, err
	}

	if err := k.Hooks().AfterValidatorBonded(ctx, consAddr, str); err != nil {
		return validator, err
	}

	return validator, err
}
