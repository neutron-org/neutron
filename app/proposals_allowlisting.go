package app

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	globalfeetypes "github.com/cosmos/gaia/v11/x/globalfee/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	packetforwardmiddlewaretypes "github.com/strangelove-ventures/packet-forward-middleware/v7/router/types"
)

func IsConsumerProposalAllowlisted(content govtypes.Content) bool {
	switch c := content.(type) {
	case *proposal.ParameterChangeProposal:
		return isConsumerParamChangeWhitelisted(c.Changes)
	case *upgradetypes.SoftwareUpgradeProposal,
		*upgradetypes.CancelSoftwareUpgradeProposal,
		*ibcclienttypes.ClientUpdateProposal,
		*ibcclienttypes.UpgradeProposal,
		*wasmtypes.PinCodesProposal,
		*wasmtypes.UnpinCodesProposal,
		*wasmtypes.UpdateAdminProposal,
		*wasmtypes.ClearAdminProposal:
		return true

	default:
		return false
	}
}

func isConsumerParamChangeWhitelisted(paramChanges []proposal.ParamChange) bool {
	for _, paramChange := range paramChanges {
		_, found := WhitelistedParams[paramChangeKey{Subspace: paramChange.Subspace, Key: paramChange.Key}]
		if !found {
			return false
		}
	}
	return true
}

type paramChangeKey struct {
	Subspace, Key string
}

var WhitelistedParams = map[paramChangeKey]struct{}{
	// ibc transfer
	{Subspace: ibctransfertypes.ModuleName, Key: string(ibctransfertypes.KeySendEnabled)}:    {},
	{Subspace: ibctransfertypes.ModuleName, Key: string(ibctransfertypes.KeyReceiveEnabled)}: {},
	// ica
	{Subspace: icahosttypes.SubModuleName, Key: string(icahosttypes.KeyHostEnabled)}:   {},
	{Subspace: icahosttypes.SubModuleName, Key: string(icahosttypes.KeyAllowMessages)}: {},
	// cosmwasm
	// {Subspace: wasmtypes.ModuleName, Key: string(wasmtypes.ParamStoreKeyUploadAccess)}:      {},
	// {Subspace: wasmtypes.ModuleName, Key: string(wasmtypes.ParamStoreKeyInstantiateAccess)}: {},
	// feerefunder
	// {Subspace: feerefundertypes.ModuleName, Key: string(feerefundertypes.KeyFees)}: {},
	// interchaintxs
	// {Subspace: interchaintxstypes.ModuleName, Key: string(interchaintxstypes.KeyMsgSubmitTxMaxMessages)}: {},
	// interchainqueries
	// {Subspace: interchainqueriestypes.ModuleName, Key: string(interchainqueriestypes.KeyQuerySubmitTimeout)}:  {},
	// {Subspace: interchainqueriestypes.ModuleName, Key: string(interchainqueriestypes.KeyQueryDeposit)}:        {},
	// {Subspace: interchainqueriestypes.ModuleName, Key: string(interchainqueriestypes.KeyTxQueryRemovalLimit)}: {},
	// feeburner
	// {Subspace: feeburnertypes.ModuleName, Key: string(feeburnertypes.KeyTreasuryAddress)}: {},
	// {Subspace: feeburnertypes.ModuleName, Key: string(feeburnertypes.KeyNeutronDenom)}:    {},
	// tokenfactory
	// {Subspace: tokenfactorytypes.ModuleName, Key: string(tokenfactorytypes.KeyDenomCreationFee)}:    {},
	// {Subspace: tokenfactorytypes.ModuleName, Key: string(tokenfactorytypes.KeyFeeCollectorAddress)}: {},
	// globalfee
	{Subspace: globalfeetypes.ModuleName, Key: string(globalfeetypes.ParamStoreKeyMinGasPrices)}: {},
	// cron
	// {Subspace: crontypes.ModuleName, Key: string(crontypes.KeySecurityAddress)}: {},
	// {Subspace: crontypes.ModuleName, Key: string(crontypes.KeyLimit)}:           {},
	// packet-forward-middleware
	{Subspace: packetforwardmiddlewaretypes.ModuleName, Key: string(packetforwardmiddlewaretypes.KeyFeePercentage)}: {},
}
