package app

import (
	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/errors"
	"cosmossdk.io/log"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	consumerante "github.com/cosmos/interchain-security/v5/app/consumer/ante"
	ibcconsumerkeeper "github.com/cosmos/interchain-security/v5/x/ccv/consumer/keeper"
	feemarketante "github.com/skip-mev/feemarket/x/feemarket/ante"

	globalfeeante "github.com/neutron-org/neutron/v5/x/globalfee/ante"
	globalfeekeeper "github.com/neutron-org/neutron/v5/x/globalfee/keeper"
)

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// channel keeper.
type HandlerOptions struct {
	ante.HandlerOptions

	BankKeeper            bankkeeper.Keeper
	AccountKeeper         feemarketante.AccountKeeper
	IBCKeeper             *ibckeeper.Keeper
	ConsumerKeeper        ibcconsumerkeeper.Keeper
	GlobalFeeKeeper       globalfeekeeper.Keeper
	WasmConfig            *wasmTypes.WasmConfig
	TXCounterStoreService corestoretypes.KVStoreService
	FeeMarketKeeper       feemarketante.FeeMarketKeeper
}

func NewAnteHandler(options HandlerOptions, logger log.Logger) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errors.Wrap(sdkerrors.ErrLogic, "account keeper is required for AnteHandler")
	}
	if options.BankKeeper == nil {
		return nil, errors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if options.SignModeHandler == nil {
		return nil, errors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}
	if options.WasmConfig == nil {
		return nil, errors.Wrap(sdkerrors.ErrLogic, "wasm config is required for ante builder")
	}
	if options.TXCounterStoreService == nil {
		return nil, errors.Wrap(sdkerrors.ErrLogic, "tx counter store service is required for ante builder")
	}

	if options.FeeMarketKeeper == nil {
		return nil, errors.Wrap(sdkerrors.ErrLogic, "feemarket keeper is required for ante builder")
	}

	sigGasConsumer := options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(),
		wasmkeeper.NewLimitSimulationGasDecorator(options.WasmConfig.SimulationGasLimit), // after setup context to enforce limits early
		wasmkeeper.NewCountTXDecorator(options.TXCounterStoreService),
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		consumerante.NewDisabledModulesDecorator("/cosmos.evidence", "/cosmos.slashing"),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		feemarketante.NewFeeMarketCheckDecorator(
			options.AccountKeeper,
			options.BankKeeper,
			options.FeegrantKeeper,
			options.FeeMarketKeeper,
			NewFeeDecoratorWithSwitch(options),
		),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
	}

	// Don't delete it even if IDE tells you so.
	// This constant depends on build tag.
	if !SkipCcvMsgFilter {
		anteDecorators = append(anteDecorators, consumerante.NewMsgFilterDecorator(options.ConsumerKeeper))
	} else {
		logger.Error("WARNING: BUILT WITH skip_ccv_msg_filter. THIS IS NOT A PRODUCTION BUILD")
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}

// FeeDecoratorWithSwitch is a fee ante decorator which switches between globalfee ante handler
// and feemarket's one, depending on the `params.Enabled` field feemarket's module.
// If feemarket is enabled, we don't need to perform checks for min gas prices, since they are handled by feemarket
// so we switch the execution directly to feemarket ante handler
// If feemarket is disabled, we call globalfee + native cosmos fee ante handler where min gas prices will be checked
// via globalfee and then they will be deducted via native cosmos fee ante handler.
type FeeDecoratorWithSwitch struct {
	globalfeeDecorator globalfeeante.FeeDecorator
	cosmosFeeDecorator ante.DeductFeeDecorator
}

func NewFeeDecoratorWithSwitch(options HandlerOptions) FeeDecoratorWithSwitch {
	return FeeDecoratorWithSwitch{
		globalfeeDecorator: globalfeeante.NewFeeDecorator(options.GlobalFeeKeeper),
		cosmosFeeDecorator: ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
	}
}

func (d FeeDecoratorWithSwitch) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// If feemarket is disabled, we call globalfee + native cosmos fee ante handler where min gas prices will be checked
	// via globalfee and then they will be deducted via native cosmos fee ante handler.
	return d.globalfeeDecorator.AnteHandle(ctx, tx, simulate, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return d.cosmosFeeDecorator.AnteHandle(ctx, tx, simulate, next)
	})
}
