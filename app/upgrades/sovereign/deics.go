package v600

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	types2 "github.com/cosmos/interchain-security/v5/x/ccv/consumer/types"
	"io/fs"
	"path/filepath"
	"time"

	"cosmossdk.io/math"
	adminmoduletypes "github.com/cosmos/admin-module/v2/x/adminmodule/types"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v5/x/ccv/consumer/keeper"
)

const (
	ICSValoperSelfStake = 1
	UnbondingTime       = 21 * 24 * time.Hour
	// neutron1jxxfkkxd9qfjzpvjyr9h3dy7l5693kx4y0zvay
	OperatorSk1 = "neutronvaloper1jxxfkkxd9qfjzpvjyr9h3dy7l5693kx47jm4mq"
	// neutron1tedsrwal9n2qlp6j3xcs3fjz9khx7z4reep8k3
	OperatorSk2 = "neutronvaloper1tedsrwal9n2qlp6j3xcs3fjz9khx7z4rryc7s4"
	// neutron1xdlvhs2l2wq0cc3eskyxphstns3348elwzvemh
	OperatorSk3 = "neutronvaloper1xdlvhs2l2wq0cc3eskyxphstns3348el5l4qan"
)

//go:embed validators/staking
var Vals embed.FS

type StakingValidator struct {
	Valoper string         `json:"valoper"`
	PK      ed25519.PubKey `json:"pk"`
}

func GatherStakingMsgs() ([]types.MsgCreateValidator, error) {
	msgs := make([]types.MsgCreateValidator, 0)
	errWalk := fs.WalkDir(Vals, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := Vals.ReadFile(path)
		if err != nil {
			return err
		}
		skval := StakingValidator{}
		err = json.Unmarshal(data, &skval)
		if err != nil {
			return err
		}
		msg := StakingValMsg(filepath.Base(path), 1000000, skval.Valoper, skval.PK)
		msgs = append(msgs, msg)
		return nil
	})
	return msgs, errWalk
}

func StakingValMsg(moniker string, stake int64, valoper string, pk ed25519.PubKey) types.MsgCreateValidator {
	pubkey, err := codectypes.NewAnyWithValue(&pk)
	if err != nil {
		panic(err)
	}
	return types.MsgCreateValidator{
		Description: types.Description{
			Moniker:         moniker,
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
		MinSelfDelegation: math.NewInt(1),
		DelegatorAddress:  "",
		// WARN: Operator must have enough funds
		ValidatorAddress: valoper,
		Pubkey:           pubkey,
		Value: sdk.Coin{
			Denom:  "untrn",
			Amount: math.NewInt(stake),
		},
	}
}

func MoveICSToStaking(ctx sdk.Context, sk stakingkeeper.Keeper, bk bankkeeper.Keeper, consumerValidators []types2.CrossChainValidator) error {
	srv := stakingkeeper.NewMsgServerImpl(&sk)
	DAOaddr, err := sdk.AccAddressFromBech32(MainDAOContractAddress)
	if err != nil {
		return err
	}

	// Add all ICS validators to staking module
	for i, v := range consumerValidators {
		// funding ICS valopers from DAO to stake a coin
		err := bk.SendCoins(ctx, DAOaddr, v.GetAddress(), sdk.NewCoins(sdk.Coin{
			Denom:  "untrn",
			Amount: math.NewInt(ICSValoperSelfStake),
		}))
		if err != nil {
			return err
		}

		valoperAddr, err := bech32.ConvertAndEncode("neutronvaloper", v.GetAddress())
		if err != nil {
			return err
		}
		_, err = srv.CreateValidator(ctx, &types.MsgCreateValidator{
			Description: types.Description{
				Moniker:         fmt.Sprintf("ics %d", i),
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
			MinSelfDelegation: math.NewInt(1),
			// WARN: valoper must have enough funds to selfbond
			ValidatorAddress: valoperAddr,
			Pubkey:           v.GetPubkey(),
			Value: sdk.Coin{
				Denom:  "untrn",
				Amount: math.NewInt(ICSValoperSelfStake),
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
		// add validator to active set to remove his from endblocker the same block
		// validator will be kicked out from active set due to the fact, voting power is calculated as `staked_amount/1_000_000`
		// staked amount for ICS validator is 1untrn => vp = 0
		_, err = bondValidator(ctx, sk, savedVal)
		if err != nil {
			return err
		}
	}

	coins := sdk.NewCoins(sdk.NewCoin("untrn", math.NewInt(int64(len(consumerValidators)*ICSValoperSelfStake))))
	// since we forced to set bond status for ics validators during the upgrade, we have to move ICS staked funds from NotBondedPoolName to BondedPoolName
	err = bk.SendCoinsFromModuleToModule(ctx, types.NotBondedPoolName, types.BondedPoolName, coins)
	return nil
}

func DeICS(ctx sdk.Context, sk stakingkeeper.Keeper, consumerKeeper ccvconsumerkeeper.Keeper, bk bankkeeper.Keeper) error {
	srv := stakingkeeper.NewMsgServerImpl(&sk)
	consumerValidators := consumerKeeper.GetAllCCValidator(ctx)
	// msgs to create new staking validators
	newValMsgs, err := GatherStakingMsgs()
	if err != nil {
		return err
	}

	params := types.Params{
		UnbondingTime: UnbondingTime,
		// During migration MaxValidators MUST be >= all the validators number, old and new ones.
		// i.e. chain managed by 150 ICS validators, and we are switching to 70 STAKING, MaxValidators MUST be at least 220,
		// otherwise panic during staking begin blocker happens
		// It's allowed to checge the value at the very next block
		MaxValidators:     uint32(len(consumerValidators) + len(newValMsgs)),
		MaxEntries:        100,
		HistoricalEntries: 100,
		BondDenom:         "untrn",
		MinCommissionRate: math.LegacyMustNewDecFromStr("0.0"),
	}

	_, err = srv.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
		Params:    params,
	})
	if err != nil {
		return err
	}

	err = MoveICSToStaking(ctx, sk, bk, consumerValidators)
	if err != nil {
		return err
	}

	for _, msg := range newValMsgs {
		_, err = srv.CreateValidator(ctx, &msg)
		if err != nil {
			return err
		}
	}

	return sk.SetLastTotalPower(ctx, math.NewInt(1))
}

// copied from staking module https://github.com/cosmos/cosmos-sdk/blob/v0.50.6/x/staking/keeper/val_state_change.go#L336
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
