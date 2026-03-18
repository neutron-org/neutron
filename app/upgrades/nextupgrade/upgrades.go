package nextupgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"

	appparams "github.com/neutron-org/neutron/v10/app/params"
	cronkeeper "github.com/neutron-org/neutron/v10/x/cron/keeper"
	"github.com/neutron-org/neutron/v10/x/cron/types"

	"github.com/neutron-org/neutron/v10/app/upgrades"
)

const (
	// MainDAOContractAddress is the address of the Neutron DAO core contract.
	MainDAOContractAddress = "neutron1suhgf5svhu4usrurvxzlgn54ksxmn8gljarjtxqnapv8kjnp4nrstdxvff"

	// IBCRateLimitsMultisig is the address that manages IBC Rate Limits contract
	IBCRateLimitsMultisig = "neutron1el2rymcsg5wxth2fz2g5l08nue3xhyj3ny5wea3yxwr9f7es8d6smmwrck"

	// IBCRateLimitsContractAddress is the address of the IBC Rate Limits contract
	IBCRateLimitsContractAddress = "neutron15aqgplxcavqhurr0g5wwtdw6025dknkqwkfh0n46gp2qjl6236cs2yd3nl"

	// RevenueModuleAccount is the address of the Revenue module account
	RevenueModuleAccount = "revenue-treasury"

	// StakingRewardsContractAddress is the address of the Staking Rewards contract
	StakingRewardsContractAddress = "neutron1gqq3c735pj6ese3yru5xr6ud0fvxgltxesygvyyzpsrt74v6yg4sgkrgwq"

	// NewMaxValidators is the new maximum number of validators
	NewMaxValidators = 1

	// PuppeteerContractAddress is the address of the Drop's Puppeteer Contract.
	// It owns all delegations including the DAO funds in Drop.
	PuppeteerContractAddress = "neutron17jsl4t4hhaw37tnhenskrfntm7mv44wzjr3f990hx4p9r5m0gzdqquhtd3"

	// PuppeteerAdmin is an admin address of the Drop's puppeteer contract
	PuppeteerAdmin = "TODO"

	// ProxyContractCodeID is the code id of the auth proxy contract code
	ProxyContractCodeID = 0o0000 // TODO

	// UndelegationsManagerContract is the address of the undelegations manager contract
	UndelegationsManagerContract = "TODO"
)

// NewValidatorSet is the target set of validators the DAO funds will be redelegated to.
// TODO: fill in real validator addresses before deployment.
var NewValidatorSet = []string{"neutronvaloper1pfklq7pcazum67hackwxr70znp09fr54q9nnva"}

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

		ctx.Logger().Info("Configuring parameters for new modules...")

		err = setDefaultParams(ctx, keepers)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Migration {nextupgrade} applied")
		return vm, nil
	}
}

// setDefaultParams sets default parameters for gov, mint, and distribution modules
func setDefaultParams(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) error {
	govparams := govtypesv1.NewParams(
		sdk.NewCoins(sdk.NewCoin(appparams.DefaultDenom, math.NewInt(1_000_000_000_000))),
		sdk.NewCoins(sdk.NewCoin(appparams.DefaultDenom, math.NewInt(1_000_000_000_000))),
		7*24*time.Hour,
		14*24*time.Hour,
		3*24*time.Hour,
		math.LegacyNewDecWithPrec(45, 2).String(),
		math.LegacyNewDecWithPrec(5, 1).String(),
		math.LegacyNewDecWithPrec(67, 2).String(),
		math.LegacyNewDecWithPrec(33, 2).String(),
		math.LegacyOneDec().String(),
		math.LegacyZeroDec().String(),
		"",
		false,
		false,
		true,
		math.LegacyMustNewDecFromStr("0.1").String(),
	)

	if err := keepers.GovKeeper.Params.Set(ctx, govparams); err != nil {
		return err
	}
	ctx.Logger().Info("Set default parameters for gov module")

	// Set default parameters for mint module
	// TODO: finalize BlocksPerYear
	mintParams := minttypes.DefaultParams()
	mintParams.MintDenom = appparams.DefaultDenom
	mintParams.InflationMax = math.LegacyNewDecWithPrec(30, 2)
	mintParams.InflationMin = math.LegacyNewDecWithPrec(1, 2)
	mintParams.InflationRateChange = math.LegacyNewDecWithPrec(2, 1)
	mintParams.GoalBonded = math.LegacyNewDecWithPrec(67, 2)
	if err := keepers.MintKeeper.Params.Set(ctx, mintParams); err != nil {
		return err
	}
	ctx.Logger().Info("Set default parameters for mint module")

	// Set default parameters for distribution module
	distrParams := distributiontypes.DefaultParams()
	distrParams.CommunityTax = math.LegacyZeroDec()
	if err := keepers.DistributionKeeper.Params.Set(ctx, distrParams); err != nil {
		return err
	}
	ctx.Logger().Info("Set default parameters for distribution module")

	ctx.Logger().Info("Initializing distribution state from staking (validators + all delegations)")
	if err := InitializeDistributionStateFromStaking(ctx, keepers.DistributionKeeper, keepers.StakingKeeper); err != nil {
		return err
	}
	ctx.Logger().Info("Done.")

	// revoke DAO and security multisig roles from IBC rate limits contract and grant root role to the gov module
	ctx.Logger().Info("Revoking DAO and security multisig roles from IBC rate limits contract and granting root role to the gov module")
	if err := IBCRateLimitsChangeRoles(ctx, keepers.WasmKeeper); err != nil {
		return err
	}
	ctx.Logger().Info("Done.")

	ctx.Logger().Info("Setting up Feemarket params")
	if err := SetupFeeMarket(ctx, keepers.FeeMarketKeeper); err != nil {
		return err
	}
	ctx.Logger().Info("Done.")

	ctx.Logger().Info("Burning funds from revenue treasury and staking rewards contract")
	if err := BurnFunds(ctx, keepers.BankKeeper); err != nil {
		return err
	}
	ctx.Logger().Info("Done.")

	ctx.Logger().Info("Setting up staking module")
	if err := SetupStaking(ctx, keepers.StakingKeeper); err != nil {
		return err
	}
	ctx.Logger().Info("Done.")

	ctx.Logger().Info("Redelegating DAO funds to new validator set")
	if err := RedelegateDaoFunds(ctx, keepers.AccountKeeper, keepers.StakingKeeper); err != nil {
		return err
	}
	ctx.Logger().Info("Done.")

	ctx.Logger().Info("Migrating puppeteer contract to auth proxy")
	if err := MigratePuppeteer(ctx, keepers.WasmKeeper); err != nil {
		return err
	}
	ctx.Logger().Info("Done.")

	ctx.Logger().Info("Registering cron schedules")
	if err := RegisterCronSchedules(ctx, &keepers.CronKeeper); err != nil {
		return err
	}
	ctx.Logger().Info("Done.")

	return nil
}

// InitializeDistributionStateFromStaking initializes distribution module state for all existing
// validators and delegations by calling the distribution module's own hooks. This is equivalent
// to what the staking module does during normal operation when validators are created and
// delegations are modified, ensuring correct state (historical rewards, reference counts,
// delegator starting info) without manual bookkeeping.
func InitializeDistributionStateFromStaking(ctx sdk.Context, dk distributionkeeper.Keeper, sk *stakingkeeper.Keeper) error {
	hooks := dk.Hooks()

	// 1) Initialize distribution state for each validator via AfterValidatorCreated.
	// This sets up: historical rewards (period 0, refcount 1), current rewards (period 1),
	// accumulated commission, and outstanding rewards.
	validators, err := sk.GetAllValidators(ctx)
	if err != nil {
		return fmt.Errorf("getting all validators: %w", err)
	}
	for _, val := range validators {
		valAddr, err := sdk.ValAddressFromBech32(val.GetOperator())
		if err != nil {
			return fmt.Errorf("invalid validator operator %s: %w", val.GetOperator(), err)
		}
		if err := hooks.AfterValidatorCreated(ctx, valAddr); err != nil {
			return fmt.Errorf("AfterValidatorCreated for %s: %w", val.GetOperator(), err)
		}
	}

	// 2) Initialize delegator starting info for each delegation, mirroring the staking
	// module's hook sequence for new delegations:
	//   - BeforeDelegationCreated: increments the validator period
	//   - AfterDelegationModified: calls initializeDelegation which increments the period
	//     again, sets DelegatorStartingInfo, and manages reference counts
	var (
		delegationCount int
		iterErr         error
	)
	err = sk.IterateAllDelegations(ctx, func(delegation stakingtypes.Delegation) bool {
		if iterErr != nil {
			return true
		}
		delAddr, err := sdk.AccAddressFromBech32(delegation.DelegatorAddress)
		if err != nil {
			iterErr = fmt.Errorf("invalid delegator %s: %w", delegation.DelegatorAddress, err)
			return true
		}
		valAddr, err := sdk.ValAddressFromBech32(delegation.ValidatorAddress)
		if err != nil {
			iterErr = fmt.Errorf("invalid validator %s: %w", delegation.ValidatorAddress, err)
			return true
		}
		if err := hooks.BeforeDelegationCreated(ctx, delAddr, valAddr); err != nil {
			iterErr = fmt.Errorf("BeforeDelegationCreated for del=%s val=%s: %w",
				delegation.DelegatorAddress, delegation.ValidatorAddress, err)
			return true
		}
		if err := hooks.AfterDelegationModified(ctx, delAddr, valAddr); err != nil {
			iterErr = fmt.Errorf("AfterDelegationModified for del=%s val=%s: %w",
				delegation.DelegatorAddress, delegation.ValidatorAddress, err)
			return true
		}
		delegationCount++
		return false
	})
	if err != nil {
		return fmt.Errorf("iterating delegations: %w", err)
	}
	if iterErr != nil {
		return iterErr
	}

	ctx.Logger().Info("Initialized distribution from staking via hooks",
		"validators", len(validators), "delegations", delegationCount)
	return nil
}

func MigratePuppeteer(ctx sdk.Context, wk *wasmkeeper.Keeper) error {
	wasmSrv := wasmkeeper.NewMsgServerImpl(wk)
	if _, err := wasmSrv.MigrateContract(ctx, &wasmTypes.MsgMigrateContract{
		Sender:   PuppeteerAdmin,
		Contract: PuppeteerContractAddress,
		CodeID:   ProxyContractCodeID,
		Msg: []byte(fmt.Sprintf(`
			{
				"owner": "%s"
			}
			`, UndelegationsManagerContract)),
	}); err != nil {
		return err
	}

	return nil
}

func RegisterCronSchedules(ctx sdk.Context, ck *cronkeeper.Keeper) error {
	if err := ck.AddSchedule(ctx, "undelegations manager contract tick & burn", 1,
		[]types.MsgExecuteContract{
			{
				Contract: UndelegationsManagerContract,
				Msg:      `{"tick": {}}`,
			},
			{
				Contract: UndelegationsManagerContract,
				Msg:      `{"burn": {}}`,
			},
		},
		types.ExecutionStage_EXECUTION_STAGE_BEGIN_BLOCKER); err != nil {
		return err
	}

	return nil
}

func IBCRateLimitsChangeRoles(ctx sdk.Context, wk *wasmkeeper.Keeper) error {
	wasmSrv := wasmkeeper.NewMsgServerImpl(wk)

	grantGovRole, err := json.Marshal(IBCRateLimitsExecuteMessage{
		GrantRole: &GrantRole{
			Signer: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			Roles: []string{
				"AddRateLimit",
				"RemoveRateLimit",
				"ResetPathQuota",
				"EditPathQuota",
				"GrantRole",
				"RevokeRole",
				"RemoveMessage",
				"SetTimelockDelay",
				"ManageDenomRestrictions",
			},
		},
	})
	if err != nil {
		return err
	}
	if _, err = wasmSrv.ExecuteContract(ctx, &wasmTypes.MsgExecuteContract{
		Sender:   MainDAOContractAddress,
		Contract: IBCRateLimitsContractAddress,
		Msg:      grantGovRole,
		Funds:    nil,
	}); err != nil {
		return err
	}

	revokeManagerRole, err := json.Marshal(IBCRateLimitsExecuteMessage{
		RevokeRole: &RevokeRole{
			Signer: IBCRateLimitsMultisig,
			Roles: []string{
				"AddRateLimit",
				"RemoveRateLimit",
				"ResetPathQuota",
				"EditPathQuota",
				"ManageDenomRestrictions",
			},
		},
	})
	if err != nil {
		return err
	}
	if _, err = wasmSrv.ExecuteContract(ctx, &wasmTypes.MsgExecuteContract{
		Sender:   MainDAOContractAddress,
		Contract: IBCRateLimitsContractAddress,
		Msg:      revokeManagerRole,
		Funds:    nil,
	}); err != nil {
		return err
	}

	revokeDAORole, err := json.Marshal(IBCRateLimitsExecuteMessage{
		RevokeRole: &RevokeRole{
			Signer: MainDAOContractAddress,
			Roles: []string{
				"AddRateLimit",
				"RemoveRateLimit",
				"ResetPathQuota",
				"EditPathQuota",
				"GrantRole",
				"RevokeRole",
				"RemoveMessage",
				"SetTimelockDelay",
				"ManageDenomRestrictions",
			},
		},
	})
	if err != nil {
		return err
	}
	if _, err = wasmSrv.ExecuteContract(ctx, &wasmTypes.MsgExecuteContract{
		Sender:   MainDAOContractAddress,
		Contract: IBCRateLimitsContractAddress,
		Msg:      revokeDAORole,
		Funds:    nil,
	}); err != nil {
		return err
	}

	return nil
}

func SetupFeeMarket(ctx context.Context, fk *feemarketkeeper.Keeper) error {
	params, err := fk.GetParams(sdk.UnwrapSDKContext(ctx))
	if err != nil {
		return err
	}

	params.SendTipToProposer = true
	err = fk.SetParams(sdk.UnwrapSDKContext(ctx), params)
	if err != nil {
		return err
	}

	return nil
}

func BurnFunds(ctx sdk.Context, bk bankkeeper.Keeper) error {
	revenueBalance := bk.GetBalance(ctx, authtypes.NewModuleAddress(RevenueModuleAccount), appparams.DefaultDenom)
	if revenueBalance.IsZero() {
		return fmt.Errorf("revenue treasury %s balance is not expected to be zero", RevenueModuleAccount)
	}
	if err := bk.SendCoinsFromAccountToModule(ctx, authtypes.NewModuleAddress(RevenueModuleAccount), "wasm", sdk.Coins{revenueBalance}); err != nil {
		return fmt.Errorf("failed to send coins from revenue treasury: %w", err)
	}

	rewardsBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(StakingRewardsContractAddress), appparams.DefaultDenom)
	if rewardsBalance.IsZero() {
		return fmt.Errorf("staking rewards contract %s balance is not expected to be zero", StakingRewardsContractAddress)
	}
	if err := bk.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(StakingRewardsContractAddress), "wasm", sdk.Coins{rewardsBalance}); err != nil {
		return fmt.Errorf("failed to send coins from staking rewards contract: %w", err)
	}

	if err := bk.BurnCoins(ctx, "wasm", sdk.NewCoins(revenueBalance)); err != nil {
		return fmt.Errorf("failed to burn revenue treasury entire balance: %w", err)
	}
	ctx.Logger().Info("Burned revenue treasury entire balance", "amount", revenueBalance)

	if err := bk.BurnCoins(ctx, "wasm", sdk.NewCoins(rewardsBalance)); err != nil {
		return fmt.Errorf("failed to burn staking rewards entire balance: %w", err)
	}
	ctx.Logger().Info("Burned staking rewards contract entire balance", "amount", rewardsBalance)

	return nil
}

func SetupStaking(ctx sdk.Context, sk *stakingkeeper.Keeper) error {
	stakingParams, err := sk.GetParams(ctx)
	if err != nil {
		return fmt.Errorf("failed to get staking module params: %w", err)
	}

	ctx.Logger().Info("Setting up staking module params with max_validators updated", "max_validators", NewMaxValidators)
	stakingParams.MaxValidators = NewMaxValidators
	if err := sk.SetParams(ctx, stakingParams); err != nil {
		return fmt.Errorf("failed to set staking module params: %w", err)
	}

	return nil
}

func RedelegateDaoFunds(ctx sdk.Context, a authkeeper.AccountKeeper, sk *stakingkeeper.Keeper) error {
	delegatorAddr := sdk.MustAccAddressFromBech32(PuppeteerContractAddress)
	delegations, err := sk.GetAllDelegatorDelegations(ctx, delegatorAddr)
	if err != nil {
		return fmt.Errorf("failed to get all delegator delegations: %w", err)
	}

	delegationsResponses, err := delegationsToDelegationResponses(ctx, a, sk, delegations)
	if err != nil {
		return fmt.Errorf("failed to get delegations responses: %w", err)
	}

	valAddresses := make([]sdk.ValAddress, len(NewValidatorSet))

	for i, addr := range NewValidatorSet {
		valAddresses[i], err = sdk.ValAddressFromBech32(addr)
		if err != nil {
			return err
		}
	}

	redelegationsMsgs := calcRedelegations(delegationsResponses, valAddresses, appparams.DefaultDenom)

	stakingMsgServer := stakingkeeper.NewMsgServerImpl(sk)
	for _, msg := range redelegationsMsgs {
		_, err := stakingMsgServer.BeginRedelegate(ctx, &msg)
		if err != nil {
			return fmt.Errorf("failed to redelegate from %s to %s: %w", msg.ValidatorSrcAddress, msg.ValidatorDstAddress, err)
		}
	}

	return nil
}

// calcRedelegations computes how to redistribute existing delegations across a new validator set
// so that each new validator receives an even distribution of tokens. When the total tokens do not
// divide evenly, the last new validator absorbs the remainder.
func calcRedelegations(
	delegations []stakingtypes.DelegationResponse,
	newValidators []sdk.ValAddress,
	denom string,
) []stakingtypes.MsgBeginRedelegate {
	redelegationsMsgs := make([]stakingtypes.MsgBeginRedelegate, 0)

	newValidatorsStack := NewStack()
	for _, val := range newValidators {
		newValidatorsStack.Push(val)
	}

	// positive means we must redelegate from this validator
	// negative means we must delegate to this validator
	DebitCredit := make(map[string]math.Int)
	totalDelegatedAmount := math.NewInt(0)
	for _, delegation := range delegations {
		totalDelegatedAmount = totalDelegatedAmount.Add(delegation.Balance.Amount)
		dc := DebitCredit[delegation.Delegation.ValidatorAddress]
		if dc.IsNil() {
			dc = math.ZeroInt()
		}
		DebitCredit[delegation.Delegation.ValidatorAddress] = dc.Add(delegation.Balance.Amount)
	}
	amountPerValidator := totalDelegatedAmount.Quo(math.NewInt(int64(len(newValidators))))
	for _, val := range newValidators {
		dc := DebitCredit[val.String()]
		if dc.IsNil() {
			dc = math.ZeroInt()
		}
		DebitCredit[val.String()] = dc.Sub(amountPerValidator)
	}
	// allocate the remainder as the last validator’s debt
	reminder := totalDelegatedAmount.Mod(math.NewInt(int64(len(newValidators))))
	DebitCredit[newValidators[len(newValidators)-1].String()] = DebitCredit[newValidators[len(newValidators)-1].String()].Sub(reminder)

	delIdx := 0
	for delIdx < len(delegations) {
		delegation := delegations[delIdx]
		newVal := newValidatorsStack.Peek()
		if newVal == nil {
			break
		}
		recvVal := newVal.String()

		remaining := DebitCredit[delegation.Delegation.ValidatorAddress]
		needed := DebitCredit[recvVal].Neg()
		if !needed.IsPositive() {
			// new validator full
			newValidatorsStack.Pop()
			continue
		}
		if !remaining.IsPositive() {
			// no more tokens to redelegate from this validator
			delIdx++
			continue
		}
		if delegation.Delegation.ValidatorAddress == newVal.String() {
			if delIdx == len(delegations)-1 && newValidatorsStack.Size() == 1 {
				// if only one delegation and one new validator, and they are "the same validator",
				// just exit the loop
				break
			}
			// push new validator to the end of queue to not delegate to itself and process later
			newValidatorsStack.Pop()
			newValidatorsStack.Push(newVal)
			continue
		}

		var take math.Int
		if remaining.LTE(needed) {
			take = remaining
		} else {
			take = needed
		}

		redelegationsMsgs = append(redelegationsMsgs, stakingtypes.MsgBeginRedelegate{
			DelegatorAddress:    delegation.Delegation.DelegatorAddress,
			ValidatorSrcAddress: delegation.Delegation.ValidatorAddress,
			ValidatorDstAddress: newVal.String(),
			// TODO: check can we just truncate shares to tokens?
			Amount: sdk.NewCoin(denom, take),
		})
		DebitCredit[delegation.Delegation.ValidatorAddress] = DebitCredit[delegation.Delegation.ValidatorAddress].Sub(take)
		DebitCredit[recvVal] = DebitCredit[recvVal].Add(take)
	}

	return redelegationsMsgs
}

// copied from staking keeper
func delegationToDelegationResponse(ctx context.Context, a authkeeper.AccountKeeper, k *stakingkeeper.Keeper, del stakingtypes.Delegation) (stakingtypes.DelegationResponse, error) {
	valAddr, err := k.ValidatorAddressCodec().StringToBytes(del.GetValidatorAddr())
	if err != nil {
		return stakingtypes.DelegationResponse{}, err
	}

	val, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		return stakingtypes.DelegationResponse{}, err
	}

	_, err = a.AddressCodec().StringToBytes(del.DelegatorAddress)
	if err != nil {
		return stakingtypes.DelegationResponse{}, err
	}

	bondDenom, err := k.BondDenom(ctx)
	if err != nil {
		return stakingtypes.DelegationResponse{}, err
	}

	return stakingtypes.NewDelegationResp(
		del.DelegatorAddress,
		del.GetValidatorAddr(),
		del.Shares,
		sdk.NewCoin(bondDenom, val.TokensFromShares(del.Shares).TruncateInt()),
	), nil
}

func delegationsToDelegationResponses(ctx context.Context, a authkeeper.AccountKeeper, k *stakingkeeper.Keeper, delegations stakingtypes.Delegations) (stakingtypes.DelegationResponses, error) {
	resp := make(stakingtypes.DelegationResponses, len(delegations))

	for i, del := range delegations {
		delResp, err := delegationToDelegationResponse(ctx, a, k, del)
		if err != nil {
			return nil, err
		}

		resp[i] = delResp
	}

	return resp, nil
}
