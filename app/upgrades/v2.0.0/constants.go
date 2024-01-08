package v200

import (
	"github.com/cosmos/cosmos-sdk/store/types"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"

	dextypes "github.com/neutron-org/neutron/v2/x/dex/types"
	"github.com/neutron-org/neutron/v2/app/upgrades"
	auctiontypes "github.com/skip-mev/block-sdk/x/auction/types"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName     = "v2.0.0"
	AtomDenom       = "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9"
	AxelarUsdcDenom = "ibc/F082B65C88E4B6D5EF1DB243CDA1D331D002759E938A0F5CD3FFDC5D53B3E349"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: types.StoreUpgrades{
		Added: []string{
			consensusparamtypes.ModuleName,
			crisistypes.ModuleName,
			dextypes.ModuleName,
			auctiontypes.ModuleName,
		},
	},
}
