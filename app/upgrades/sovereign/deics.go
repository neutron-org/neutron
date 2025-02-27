package v600

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	types2 "github.com/cosmos/interchain-security/v5/x/ccv/consumer/types"
	"github.com/neutron-org/neutron/v5/app/params"
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
	// set of constants defines self delegation amount for newly created validator
	// ICS and Sovereign ones
	SovereignMinSelfDelegation = 1_000_000
	SovereignSelfStake         = 1_000_000
	ICSMinSelfDelegation       = 1
	ICSSelfStake               = 1

	UnbondingTime = 21 * 24 * time.Hour
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
		msg := StakingValMsg(filepath.Base(path), SovereignSelfStake, skval.Valoper, skval.PK)
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
		MinSelfDelegation: math.NewInt(SovereignMinSelfDelegation),
		DelegatorAddress:  "",
		// WARN: Operator must have enough funds
		ValidatorAddress: valoper,
		Pubkey:           pubkey,
		Value: sdk.Coin{
			Denom:  params.DefaultDenom,
			Amount: math.NewInt(stake),
		},
	}
}

// MoveICSToStaking crates CCV validators in staking module, forces to change status to bonded
// to generate valsetupdate with 0 vp the same block
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
			Denom:  params.DefaultDenom,
			Amount: math.NewInt(ICSSelfStake),
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
			MinSelfDelegation: math.NewInt(ICSMinSelfDelegation),
			// WARN: valoper must have enough funds to selfbond
			ValidatorAddress: valoperAddr,
			Pubkey:           v.GetPubkey(),
			Value: sdk.Coin{
				Denom:  params.DefaultDenom,
				Amount: math.NewInt(ICSSelfStake),
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

	coins := sdk.NewCoins(sdk.NewCoin(params.DefaultDenom, math.NewInt(int64(len(consumerValidators)*ICSSelfStake))))
	// since we forced to set bond status for ics validators during the upgrade, we have to move ICS staked funds from NotBondedPoolName to BondedPoolName
	return bk.SendCoinsFromModuleToModule(ctx, types.NotBondedPoolName, types.BondedPoolName, coins)
}

// DeICS - does the deics, The whole point of the method is force staking module to remove ICS validators
// and add STAKING(sovereign) ones, by generating valsetupdates (https://github.com/cosmos/cosmos-sdk/blob/v0.50.9/x/staking/keeper/val_state_change.go#L269)
// We add STAKING and ICS to staking module in a special way.
// STAKING added in natural staking way, just submit `MsgCreateValidator` msg with vp >=1.
// ICS added with the message (with vp = 0). And to force staking to remove the ICS validators the same block,
// we force validators to join "active" set by bonding them with `bondValidator` method and move stake from nonbonded pool to bonded.
//
// The same block migration happened, in a staking module we have STAKING validators in `unbonded` state with vp>=1 and ICS in `bonded` state with vp==0.
//
// In the endblocker, staking module iterates over all validators in order of decreasing vp. https://github.com/cosmos/cosmos-sdk/blob/v0.50.9/x/staking/keeper/val_state_change.go#L155
// STAKING validators change the status from unbonded to bonded https://github.com/cosmos/cosmos-sdk/blob/v0.50.9/x/staking/keeper/val_state_change.go#L174,
// and valupdete generated https://github.com/cosmos/cosmos-sdk/blob/v0.50.9/x/staking/keeper/val_state_change.go#L202. STAKING validators removed from list `last` https://github.com/cosmos/cosmos-sdk/blob/v0.50.9/x/staking/keeper/val_state_change.go#L209 which at the start of endblocker contains STAKING+ICS validators.
// when the iterator reaches first ICS validator, it stopes processing https://github.com/cosmos/cosmos-sdk/blob/v0.50.9/x/staking/keeper/val_state_change.go#L168
// By that time only ICS left in `last` list
// `last` list becomes `noLongerBonded` https://github.com/cosmos/cosmos-sdk/blob/v0.50.9/x/staking/keeper/val_state_change.go#L215
// and valupdate with zero power (to remove from cometbft) generated https://github.com/cosmos/cosmos-sdk/blob/v0.50.9/x/staking/keeper/val_state_change.go#L235
// Right after migration (preblocker) we have
// ICS - bonded, STAKING - unbonded
// At the end of the block, after endblocker
// ICS - unbonded, STAKING - bonded
// cometbft valset gets updated with a lag
// upgrade + `next block` still sogned by ICS validators
// upgrade + 2 the first block signed by STAKING validators.
func DeICS(ctx sdk.Context, sk stakingkeeper.Keeper, consumerKeeper ccvconsumerkeeper.Keeper, bk bankkeeper.Keeper) error {
	srv := stakingkeeper.NewMsgServerImpl(&sk)
	consumerValidators := consumerKeeper.GetAllCCValidator(ctx)
	// msgs to create new staking validators
	newValMsgs, err := GatherStakingMsgs()
	if err != nil {
		return err
	}

	p := types.Params{
		UnbondingTime: UnbondingTime,
		// During migration MaxValidators MUST be >= all the validators number, old and new ones.
		// i.e. chain managed by 150 ICS validators, and we are switching to 70 STAKING, MaxValidators MUST be at least 220,
		// otherwise panic during staking begin blocker happens
		// It's allowed to change the value at the very next block
		MaxValidators:     uint32(len(consumerValidators) + len(newValMsgs)),
		MaxEntries:        100,
		HistoricalEntries: 100,
		BondDenom:         params.DefaultDenom,
		MinCommissionRate: math.LegacyMustNewDecFromStr("0.0"),
	}

	_, err = srv.UpdateParams(ctx, &types.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
		Params:    p,
	})
	if err != nil {
		return err
	}

	err = MoveICSToStaking(ctx, sk, bk, consumerValidators)
	if err != nil {
		return err
	}

	DAOaddr, err := sdk.AccAddressFromBech32(MainDAOContractAddress)
	if err != nil {
		return err
	}

	for _, msg := range newValMsgs {
		valAddr, err := sdk.GetFromBech32(msg.ValidatorAddress, "neutronvaloper")
		if err != nil {
			return err
		}

		// prefund validator to make selfbond
		err = bk.SendCoins(ctx, DAOaddr, valAddr, sdk.NewCoins(sdk.Coin{
			Denom:  params.DefaultDenom,
			Amount: math.NewInt(SovereignSelfStake),
		}))
		if err != nil {
			return err
		}

		_, err = srv.CreateValidator(ctx, &msg)
		if err != nil {
			return err
		}
	}

	return nil
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
