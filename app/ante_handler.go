package app

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibcante "github.com/cosmos/ibc-go/v4/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v4/modules/core/keeper"
	consumerante "github.com/cosmos/interchain-security/app/consumer/ante"
	ibcconsumerkeeper "github.com/cosmos/interchain-security/x/ccv/consumer/keeper"
	"github.com/tendermint/tendermint/libs/log"

	globalfeeante "github.com/cosmos/gaia/v8/x/globalfee/ante"
)

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// channel keeper.
type HandlerOptions struct {
	ante.HandlerOptions

	IBCKeeper            *ibckeeper.Keeper
	ConsumerKeeper       ibcconsumerkeeper.Keeper
	WasmConfig           *wasmTypes.WasmConfig
	TXCounterStoreKey    sdk.StoreKey
	BypassMinFeeMsgTypes []string
	GlobalFeeSubspace    paramtypes.Subspace
}

func NewAnteHandler(options HandlerOptions, logger log.Logger) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for AnteHandler")
	}
	if options.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if options.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}
	if options.WasmConfig == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "wasm config is required for ante builder")
	}
	if options.TXCounterStoreKey == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "tx counter key is required for ante builder")
	}
	if options.GlobalFeeSubspace.Name() == "" {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "globalfee param store is required for AnteHandler")
	}

	sigGasConsumer := options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	// maxBypassMinFeeMsgGasUsage is the maximum gas usage per message
	// so that a transaction that contains only message types that can
	// bypass the minimum fee can be accepted with a zero fee.
	// For details, see gaiafeeante.NewFeeDecorator()
	const maxBypassMinFeeMsgGasUsage uint64 = 500_000 // Should be high enough because /ibc.core.client.v1.MsgUpdateClient is the most expensive message

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(),
		wasmkeeper.NewLimitSimulationGasDecorator(options.WasmConfig.SimulationGasLimit), // after setup context to enforce limits early
		wasmkeeper.NewCountTXDecorator(options.TXCounterStoreKey),
		ante.NewRejectExtensionOptionsDecorator(),
		consumerante.NewDisabledModulesDecorator("/cosmos.evidence", "/cosmos.slashing"),
		ante.NewMempoolFeeDecorator(),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		// We are providing options.GlobalFeeSubspace because we do not have staking module
		// In this case you should be sure that you implemented upgrade to set default global fee param and it SHOULD contain at least one record
		// otherwise you will get panic
		globalfeeante.NewFeeDecorator(options.BypassMinFeeMsgTypes, options.GlobalFeeSubspace, options.GlobalFeeSubspace, maxBypassMinFeeMsgGasUsage),

		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewAnteDecorator(options.IBCKeeper),
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
