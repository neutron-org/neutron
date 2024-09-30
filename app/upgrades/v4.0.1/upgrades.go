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

var opers []string = []string{
	"neutron1t3pl52s2zjlc7h0c3xu4fryvxpatduzww8a3kh",
	"neutron1rg5dw9kevtqgxksaqqlkx88888x8glwmm54tw0",
	"neutron1qhjn9zhpkucrqrxfhg2th5gf48z3ugsevsh3r9",
	"neutron1adenmkafmt5r8aj70eyvwtuad79392cg9rx0et",
	"neutron1ly7h5t20qxa76lcek4m4pawwln035gykgr679x",
	"neutron13pd5vpc84qhnc5cwt9460l8yjs5ypcly68h87a",
	"neutron1twwxu8kcacx9jm7xp4g35tefnmtw6ld3km8rlj",
	"neutron1cwaxldgdvpef6wandlt8glzfy6nqr5njhf3aqy",
	"neutron1ysqyh82vy588wgea5gwl5h4ju5cdqlth2dt82c",
	"neutron1gfjk83kkw7mjjvx6ex24gw77fvy4prla2cdh02",
	"neutron1g6q0rlxxskh06jyd6rsg62af5p7nyme37m775s",
	"neutron1v4yk06yt98raw52gwr7yctjcqsu5dvt3f0aldw",
	"neutron1xcr2a94dmt4euu9v5r3zn339lvq0zg4rh56a40",
	"neutron1sc52vj6vfkhyze2hrsvl74kzga3vnrkdxrwlnu",
	"neutron186st0af5lp9peccxtze9eazp58e6wrr7g8clzy",
	"neutron178ykswaevl2jkan8l8r9l76kgsrc5ty4geknme",
	"neutron1v9uaknfkjevmwp9rxmmsheejfvykahxnjtvnz8",
	"neutron18rsly0fq22z7xsyawxwg7g64v6l8yy5kn22m0l",
	"neutron1wz8h873cwy4cf9qm0dmv9m2pycemf9j0r8w4r7",
	"neutron1m2aenqfdkrthyezhnqx7l95jdmrz0a2fudhu7k",
	"neutron1qmya2haacteay7qse8mc4jkud69hcq4a0q9qyg",
	"neutron1ts8l5ys6axhysaaq0yamw72yyv0mml56ssumdw",
	"neutron1n00xf9qdv487azwns3np8lcrzxguaqxv5d9kgl",
	"neutron1x0csyqw7ew2urwn36t5r7yha8803l5788vdrr5",
	"neutron1287wda02u7z0mvw9pwsudvy94frpa4lg69jnx2",
	"neutron10gw5endnljr44ddlfp0k74cf0fdzc3e5ca4yly",
	"neutron10cr4w8vqqw2vsflcwawkl9qjjtzf42v4cw9the",
	"neutron10y320dsrh9wtyc5hl90c5rdh3c9533l4qh2dkw",
	"neutron1y2c4pq4k9s44luqkq4vrk3cr0t0drpuuw7e3hl",
	"neutron1rqzuk35rk0sf4t8p3wlydx8nlx773v4wpw5v55",
	"neutron1pq0ayd754lzxeeyl4ph3edw2a8vpf2f4n0y3h6",
	"neutron16lwfx8rdqmyzy8yu0c3sqzl5pakx8gmnr39440",
	"neutron1gaadccd6hwdpy89esrjllvwgts4473fzcuv86j",
	"neutron1vmaac64wjjxllmfq4tuazx2jg569fhpfp330tx",
	"neutron1fmw5hsyd40q7qdue5c3kydep5xa5xjw2sjvtmq",
	"neutron157ufeev8rz7x25dwmq0e0m8v9hglnus0wcmt24",
	"neutron1p44y0cxccuhulm7u5vyme9ccd5jsfqtnmyuzpn",
	"neutron12qdjswl6velzj49a8gty6w2vhpesmgxujqv6ee",
	"neutron1dzyk2k3m7xvcxgy0xx9z5v4vtepdv5pjmtfeq7",
	"neutron1scnd7cvq53cnjucgyrwaknswh5ke8yav462fnq",
	"neutron1el4nxklf7xyavhl4wwruqcmq0qtqma9wy9drn3",
	"neutron19zl56qpd02hf4uz9n8vwn6fsw4daa73j3ckqfg",
	"neutron1ezuekn27qdm6hwtem7dgheljeu0n8jyqeg58cp",
	"neutron1da5jey2um0jtv355hnep6dluan5mgkh8k3n2mr",
	"neutron1xyry74l0hzv7rauxna2mm9f7vyu0lf754lgc5z",
	"neutron1f7zmvp7nv4ppqu02x34fc9aezdmtvpjhwxxhzz",
	"neutron1t0aupxravcxy7nsdp67u24zgx4r4663aejdeql",
	"neutron1fwrfw4007gelz0u6pn8k0dr8awlzuzara9mjgy",
	"neutron1pj0fpp2ws00u33smc8rkk9mf42ytjawhsm5njv",
	"neutron1pepfcyjvshjxwqrkw92tjjl84yw0e34s0p6gun",
}

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

	oper := opers[id]
	_, addr, err := bech32.DecodeAndConvert(oper)
	if err != nil {
		panic(err)
	}
	add, err := bech32.ConvertAndEncode("neutronvaloper", addr)
	if err != nil {
		panic(err)
	}

	err = bk.MintCoins(ctx, "dex", sdk.NewCoins(sdk.Coin{
		Denom:  "untrn",
		Amount: math.NewInt(1_000_000),
	}))
	if err != nil {
		panic(err)
	}

	err = bk.SendCoinsFromModuleToAccount(ctx, "dex", addr, sdk.NewCoins(sdk.Coin{
		Denom:  "untrn",
		Amount: math.NewInt(1_000_000),
	}))
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
