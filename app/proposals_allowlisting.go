package app

import (
	upgradetypes "cosmossdk.io/x/upgrade/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	pfmtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck
	ccvconsumertypes "github.com/cosmos/interchain-security/v5/x/ccv/consumer/types"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"
	marketmaptypes "github.com/skip-mev/slinky/x/marketmap/types"

	ibcratelimittypes "github.com/neutron-org/neutron/v5/x/ibc-rate-limit/types"

	dynamicfeestypes "github.com/neutron-org/neutron/v5/x/dynamicfees/types"
	globalfeetypes "github.com/neutron-org/neutron/v5/x/globalfee/types"

	contractmanagertypes "github.com/neutron-org/neutron/v5/x/contractmanager/types"
	crontypes "github.com/neutron-org/neutron/v5/x/cron/types"
	dextypes "github.com/neutron-org/neutron/v5/x/dex/types"
	feeburnertypes "github.com/neutron-org/neutron/v5/x/feeburner/types"
	feerefundertypes "github.com/neutron-org/neutron/v5/x/feerefunder/types"
	interchainqueriestypes "github.com/neutron-org/neutron/v5/x/interchainqueries/types"
	interchaintxstypes "github.com/neutron-org/neutron/v5/x/interchaintxs/types"
	tokenfactorytypes "github.com/neutron-org/neutron/v5/x/tokenfactory/types"
)

func IsConsumerProposalAllowlisted(content govtypes.Content) bool {
	switch c := content.(type) {
	case *proposal.ParameterChangeProposal:
		return isConsumerParamChangeWhitelisted(c.Changes)
	case *ibcclienttypes.ClientUpdateProposal, //nolint:staticcheck
		*ibcclienttypes.UpgradeProposal: //nolint:staticcheck
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
		*consensustypes.MsgUpdateParams,
		*upgradetypes.MsgSoftwareUpgrade,
		*upgradetypes.MsgCancelUpgrade,
		*ibcclienttypes.MsgRecoverClient,
		*ibcclienttypes.MsgIBCSoftwareUpgrade,
		*tokenfactorytypes.MsgUpdateParams,
		*interchainqueriestypes.MsgUpdateParams,
		*interchaintxstypes.MsgUpdateParams,
		*feeburnertypes.MsgUpdateParams,
		*feerefundertypes.MsgUpdateParams,
		*crontypes.MsgUpdateParams,
		*crontypes.MsgAddSchedule,
		*crontypes.MsgRemoveSchedule,
		*contractmanagertypes.MsgUpdateParams,
		*dextypes.MsgUpdateParams,
		*banktypes.MsgUpdateParams,
		*crisistypes.MsgUpdateParams,
		*minttypes.MsgUpdateParams,
		*pfmtypes.MsgUpdateParams,
		*marketmaptypes.MsgCreateMarkets,
		*marketmaptypes.MsgUpdateMarkets,
		*marketmaptypes.MsgRemoveMarketAuthorities,
		*marketmaptypes.MsgParams,
		*authtypes.MsgUpdateParams,
		*ccvconsumertypes.MsgUpdateParams,
		*icahosttypes.MsgUpdateParams,
		*feemarkettypes.MsgParams,
		*dynamicfeestypes.MsgUpdateParams,
		*ibctransfertypes.MsgUpdateParams,
		*globalfeetypes.MsgUpdateParams,
		*ibcratelimittypes.MsgUpdateParams:
		return true
	}
	return false
}

type paramChangeKey struct {
	Subspace, Key string
}

var WhitelistedParams = map[paramChangeKey]struct{}{}
