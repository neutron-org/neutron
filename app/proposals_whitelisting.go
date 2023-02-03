package app

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icahosttypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"

	feerefundertypes "github.com/neutron-org/neutron/x/feerefunder/types"
	interchainqueriestypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	interchaintxstypes "github.com/neutron-org/neutron/x/interchaintxs/types"
)

func IsConsumerProposalWhitelisted(content govtypes.Content) bool {
	switch c := content.(type) {
	case *proposal.ParameterChangeProposal:
		return isConsumerParamChangeWhitelisted(c.Changes)
	case *govtypes.TextProposal,
		*upgradetypes.SoftwareUpgradeProposal,
		*upgradetypes.CancelSoftwareUpgradeProposal,
		*ibcclienttypes.ClientUpdateProposal,
		*ibcclienttypes.UpgradeProposal,
		*wasmtypes.StoreCodeProposal,
		*wasmtypes.InstantiateContractProposal,
		*wasmtypes.MigrateContractProposal,
		*wasmtypes.SudoContractProposal,
		*wasmtypes.ExecuteContractProposal,
		*wasmtypes.UpdateAdminProposal,
		*wasmtypes.ClearAdminProposal,
		*wasmtypes.PinCodesProposal,
		*wasmtypes.UnpinCodesProposal,
		*wasmtypes.UpdateInstantiateConfigProposal:
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
	// bank
	{Subspace: banktypes.ModuleName, Key: string(banktypes.KeySendEnabled)}: {},
	// ibc transfer
	{Subspace: ibctransfertypes.ModuleName, Key: string(ibctransfertypes.KeySendEnabled)}:    {},
	{Subspace: ibctransfertypes.ModuleName, Key: string(ibctransfertypes.KeyReceiveEnabled)}: {},
	// ica
	{Subspace: icahosttypes.SubModuleName, Key: string(icahosttypes.KeyHostEnabled)}:   {},
	{Subspace: icahosttypes.SubModuleName, Key: string(icahosttypes.KeyAllowMessages)}: {},
	// cosmwasm
	{Subspace: wasmtypes.ModuleName, Key: string(wasmtypes.ParamStoreKeyUploadAccess)}:      {},
	{Subspace: wasmtypes.ModuleName, Key: string(wasmtypes.ParamStoreKeyInstantiateAccess)}: {},
	// feerefunder
	{Subspace: feerefundertypes.ModuleName, Key: string(feerefundertypes.KeyFees)}: {},
	// interchaintxs
	{Subspace: interchaintxstypes.ModuleName, Key: string(interchaintxstypes.KeyMsgSubmitTxMaxMessages)}: {},
	// interchainqueries
	{Subspace: interchainqueriestypes.ModuleName, Key: string(interchainqueriestypes.KeyQuerySubmitTimeout)}: {},
	{Subspace: interchainqueriestypes.ModuleName, Key: string(interchainqueriestypes.KeyQueryDeposit)}:       {},
}
