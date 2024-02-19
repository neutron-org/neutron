package app

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	globalfeetypes "github.com/cosmos/gaia/v11/x/globalfee/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"github.com/neutron-org/neutron/v2/app/params"
)

// GenesisState is the genesis state of the blockchain represented here as a map of raw json
// messages key'd by a identifier string.
// The identifier is used to determine which module genesis information belongs
// to so it may be appropriately routed during init chain.
// Within this application default genesis information is retrieved from
// the ModuleBasicManager which populates json from each BasicModule
// object provided to it during init.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState(cdc codec.JSONCodec) GenesisState {
	// This ugly hack is required to alter globalfee module genesis state
	// because in current chain implementation staking module is absent which is required by globalfee module
	// and we can't use default genesis state for globalfee module.
	// If we do not alter globalfee module genesis state, then we will get panic during tests run.

	genesisState := ModuleBasics.DefaultGenesis(cdc)
	globalFeeGenesisState := globalfeetypes.GenesisState{
		Params: globalfeetypes.Params{
			MinimumGasPrices: sdk.DecCoins{
				sdk.NewDecCoinFromDec(params.DefaultDenom, sdk.MustNewDecFromStr("0")),
			},
			BypassMinFeeMsgTypes: []string{
				sdk.MsgTypeURL(&ibcchanneltypes.MsgRecvPacket{}),
				sdk.MsgTypeURL(&ibcchanneltypes.MsgAcknowledgement{}),
				sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{}),
			},
			MaxTotalBypassMinFeeMsgGasUsage: globalfeetypes.DefaultmaxTotalBypassMinFeeMsgGasUsage,
		},
	}
	globalFeeGenesisStateBytes, err := json.Marshal(globalFeeGenesisState)
	if err != nil {
		panic("cannot marshal globalfee genesis state for tests")
	}
	genesisState["globalfee"] = globalFeeGenesisStateBytes

	return genesisState
}
