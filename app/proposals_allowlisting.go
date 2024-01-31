package app

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	globalfeetypes "github.com/cosmos/gaia/v11/x/globalfee/types"
	pfmtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ccvconsumertypes "github.com/cosmos/interchain-security/v4/x/ccv/consumer/types"
	ccv "github.com/cosmos/interchain-security/v4/x/ccv/types"

	contractmanagertypes "github.com/neutron-org/neutron/v2/x/contractmanager/types"
	crontypes "github.com/neutron-org/neutron/v2/x/cron/types"
	dextypes "github.com/neutron-org/neutron/v2/x/dex/types"
	feeburnertypes "github.com/neutron-org/neutron/v2/x/feeburner/types"
	feerefundertypes "github.com/neutron-org/neutron/v2/x/feerefunder/types"
	interchainqueriestypes "github.com/neutron-org/neutron/v2/x/interchainqueries/types"
	interchaintxstypes "github.com/neutron-org/neutron/v2/x/interchaintxs/types"
	tokenfactorytypes "github.com/neutron-org/neutron/v2/x/tokenfactory/types"
)

func IsConsumerProposalAllowlisted(content govtypes.Content) bool {
	switch c := content.(type) {
	case *proposal.ParameterChangeProposal:
		return isConsumerParamChangeWhitelisted(c.Changes)
	case *ibcclienttypes.ClientUpdateProposal,
		*ibcclienttypes.UpgradeProposal:
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

// This function is designed to determine if a given message (of type sdk.Msg) belongs to
// a predefined whitelist of message types which could be executed via admin module.
func isSdkMessageWhitelisted(msg sdk.Msg) bool {
	switch msg.(type) {
	case *wasmtypes.MsgClearAdmin,
		*wasmtypes.MsgUpdateAdmin,
		*wasmtypes.MsgUpdateParams,
		*wasmtypes.MsgPinCodes,
		*wasmtypes.MsgUnpinCodes,
		*upgradetypes.MsgSoftwareUpgrade,
		*upgradetypes.MsgCancelUpgrade,
		*tokenfactorytypes.MsgUpdateParams,
		*interchainqueriestypes.MsgUpdateParams,
		*interchaintxstypes.MsgUpdateParams,
		*feeburnertypes.MsgUpdateParams,
		*feerefundertypes.MsgUpdateParams,
		*crontypes.MsgUpdateParams,
		*contractmanagertypes.MsgUpdateParams,
		*dextypes.MsgUpdateParams,
		*banktypes.MsgUpdateParams,
		*crisistypes.MsgUpdateParams,
		*minttypes.MsgUpdateParams,
		*pfmtypes.MsgUpdateParams,
		*authtypes.MsgUpdateParams:
		return true
	}
	return false
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
	// globalfee
	{Subspace: globalfeetypes.ModuleName, Key: string(globalfeetypes.ParamStoreKeyMinGasPrices)}:                    {},
	{Subspace: globalfeetypes.ModuleName, Key: string(globalfeetypes.ParamStoreKeyBypassMinFeeMsgTypes)}:            {},
	{Subspace: globalfeetypes.ModuleName, Key: string(globalfeetypes.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage)}: {},
	// ICS consumer
	{Subspace: ccvconsumertypes.ModuleName, Key: string(ccv.KeyRewardDenoms)}:                      {},
	{Subspace: ccvconsumertypes.ModuleName, Key: string(ccv.KeyEnabled)}:                           {},
	{Subspace: ccvconsumertypes.ModuleName, Key: string(ccv.KeyBlocksPerDistributionTransmission)}: {},
	{Subspace: ccvconsumertypes.ModuleName, Key: string(ccv.KeyDistributionTransmissionChannel)}:   {},
	{Subspace: ccvconsumertypes.ModuleName, Key: string(ccv.KeyProviderFeePoolAddrStr)}:            {},
	{Subspace: ccvconsumertypes.ModuleName, Key: string(ccv.KeyTransferTimeoutPeriod)}:             {},
	{Subspace: ccvconsumertypes.ModuleName, Key: string(ccv.KeyConsumerRedistributionFrac)}:        {},
	{Subspace: ccvconsumertypes.ModuleName, Key: string(ccv.KeyHistoricalEntries)}:                 {},
	{Subspace: ccvconsumertypes.ModuleName, Key: string(ccv.KeyConsumerUnbondingPeriod)}:           {},
	{Subspace: ccvconsumertypes.ModuleName, Key: string(ccv.KeySoftOptOutThreshold)}:               {},
	{Subspace: ccvconsumertypes.ModuleName, Key: string(ccv.KeyProviderRewardDenoms)}:              {},
	{Subspace: ccvconsumertypes.ModuleName, Key: string(ccv.KeyRetryDelayPeriod)}:                  {},
}
