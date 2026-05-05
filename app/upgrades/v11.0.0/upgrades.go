package v11_0_0

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"
	marketmapkeeper "github.com/skip-mev/slinky/x/marketmap/keeper"

	appparams "github.com/neutron-org/neutron/v11/app/params"

	"github.com/neutron-org/neutron/v11/app/upgrades"
)

const (
	// MainDAOContractAddress is the address of the Neutron DAO core contract.
	MainDAOContractAddress = "neutron1kvxlf27r0h7mzjqgdydqdf76dtlyvwz6u9q8tysfae53ajv8urtq4fdkvy"

	// NewMaxValidators is the new maximum number of validators
	NewMaxValidators = 11

	// PuppeteerContractAddress is the address of the Drop's Puppeteer Contract.
	// It owns all delegations including the DAO funds in Drop.
	PuppeteerContractAddress = "neutron1jc4c43n36vkx7x0ke7lvhs2386ar9q4adevzpex650ff4zp0gfyq07xuea"
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

		ctx.Logger().Info("Starting migration steps...")
		err = executeUpgradeSteps(ctx, keepers)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Migration {nextupgrade} applied")
		return vm, nil
	}
}

// executeUpgradeSteps sets default parameters for gov, mint, and distribution modules
func executeUpgradeSteps(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) error {
	ctx.Logger().Info("Configuring parameters for new modules...")
	if err := setModuleParams(ctx, keepers); err != nil {
		return err
	}
	ctx.Logger().Info("Done.")

	ctx.Logger().Info("Initializing distribution state from staking (validators + all delegations)")
	if err := InitializeDistributionStateFromStaking(ctx, keepers.DistributionKeeper, keepers.StakingKeeper); err != nil {
		return err
	}
	ctx.Logger().Info("Done.")

	ctx.Logger().Info("Setting up Feemarket params")
	if err := SetupFeeMarket(ctx, keepers.FeeMarketKeeper); err != nil {
		return err
	}
	ctx.Logger().Info("Done.")

	ctx.Logger().Info("Setting up MarketMap params")
	if err := SetupMarketMap(ctx, keepers.MarketmapKeeper); err != nil {
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

	return nil
}

func setModuleParams(ctx sdk.Context, keepers *upgrades.UpgradeKeepers) error {
	maxDepositPeriod := 3 * 24 * time.Hour
	votingPeriod := 24 * time.Hour
	expeditedVotingPeriod := 24 * time.Hour
	govparams := govtypesv1.Params{
		MinDeposit:                 sdk.NewCoins(sdk.NewCoin(appparams.DefaultDenom, math.NewInt(300_000_000_000))),
		ExpeditedMinDeposit:        sdk.NewCoins(sdk.NewCoin(appparams.DefaultDenom, math.NewInt(1_000_000_000_000))),
		MaxDepositPeriod:           &maxDepositPeriod,
		VotingPeriod:               &votingPeriod,
		ExpeditedVotingPeriod:      &expeditedVotingPeriod,
		Quorum:                     math.LegacyNewDecWithPrec(30, 2).String(),
		Threshold:                  math.LegacyNewDecWithPrec(5, 1).String(),
		ExpeditedThreshold:         math.LegacyNewDecWithPrec(67, 2).String(),
		VetoThreshold:              math.LegacyNewDecWithPrec(33, 2).String(),
		MinInitialDepositRatio:     math.LegacyNewDecWithPrec(5, 1).String(),
		ProposalCancelRatio:        math.LegacyZeroDec().String(),
		ProposalCancelDest:         "",
		BurnProposalDepositPrevote: false,
		BurnVoteQuorum:             false,
		BurnVoteVeto:               true,
		MinDepositRatio:            math.LegacyMustNewDecFromStr("0.1").String(),
	}

	if err := keepers.GovKeeper.Params.Set(ctx, govparams); err != nil {
		return err
	}
	ctx.Logger().Info("Set default parameters for gov module")

	// Set default parameters for mint module
	mintParams := minttypes.DefaultParams()
	mintParams.MintDenom = appparams.DefaultDenom
	mintParams.InflationMax = math.LegacyNewDecWithPrec(30, 2)
	mintParams.InflationMin = math.LegacyNewDecWithPrec(1, 2)
	mintParams.InflationRateChange = math.LegacyNewDecWithPrec(2, 1)
	mintParams.GoalBonded = math.LegacyNewDecWithPrec(67, 2)
	mintParams.BlocksPerYear = uint64(60 * 60 * 8760 / 1) // assuming 1s per block
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

func SetupMarketMap(ctx context.Context, mmk *marketmapkeeper.Keeper) error {
	params, err := mmk.GetParams(sdk.UnwrapSDKContext(ctx))
	if err != nil {
		return err
	}

	params.Admin = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	params.MarketAuthorities = []string{authtypes.NewModuleAddress(govtypes.ModuleName).String()}
	if err := mmk.SetParams(sdk.UnwrapSDKContext(ctx), params); err != nil {
		return err
	}

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
	newValIdx := 0
	for delIdx < len(delegations) && newValIdx < len(newValidators) {
		delegation := delegations[delIdx]
		newVal := newValidators[newValIdx].String()

		remaining := DebitCredit[delegation.Delegation.ValidatorAddress]
		needed := DebitCredit[newVal].Neg()
		if !needed.IsPositive() {
			// new validator full
			newValIdx++
			continue
		}
		if !remaining.IsPositive() {
			// no more tokens to redelegate from this validator
			delIdx++
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
			ValidatorDstAddress: newVal,
			Amount:              sdk.NewCoin(denom, take),
		})
		DebitCredit[delegation.Delegation.ValidatorAddress] = DebitCredit[delegation.Delegation.ValidatorAddress].Sub(take)
		DebitCredit[newVal] = DebitCredit[newVal].Add(take)
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
