package upgrades

import (
	store "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	adminmodulekeeper "github.com/cosmos/admin-module/v2/x/adminmodule/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	channelkeeper "github.com/cosmos/ibc-go/v10/modules/core/04-channel/keeper"
	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"
	marketmapkeeper "github.com/skip-mev/slinky/x/marketmap/keeper"

	harpoonkeeper "github.com/neutron-org/neutron/v7/x/harpoon/keeper"
	revenuekeeper "github.com/neutron-org/neutron/v7/x/revenue/keeper"

	dexkeeper "github.com/neutron-org/neutron/v7/x/dex/keeper"
	ibcratelimitkeeper "github.com/neutron-org/neutron/v7/x/ibc-rate-limit/keeper"

	dynamicfeeskeeper "github.com/neutron-org/neutron/v7/x/dynamicfees/keeper"

	contractmanagerkeeper "github.com/neutron-org/neutron/v7/x/contractmanager/keeper"
	cronkeeper "github.com/neutron-org/neutron/v7/x/cron/keeper"
	feeburnerkeeper "github.com/neutron-org/neutron/v7/x/feeburner/keeper"
	icqkeeper "github.com/neutron-org/neutron/v7/x/interchainqueries/keeper"
	tokenfactorykeeper "github.com/neutron-org/neutron/v7/x/tokenfactory/keeper"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	transferkeeper "github.com/cosmos/ibc-go/v10/modules/apps/transfer/keeper"
)

// Upgrade defines a struct containing necessary fields that a SoftwareUpgradeProposal
// must have written, in order for the state migration to go smoothly.
// An upgrade must implement this struct, and then set it in the app.go.
// The app.go will then define the handler.
type Upgrade struct {
	// Upgrade version name, for the upgrade handler, e.g. `v7`
	UpgradeName string

	// CreateUpgradeHandler defines the function that creates an upgrade handler
	CreateUpgradeHandler func(*module.Manager, module.Configurator, *UpgradeKeepers, StoreKeys, codec.Codec) upgradetypes.UpgradeHandler

	// Store upgrades, should be used for any new modules introduced, new modules deleted, or store names renamed.
	StoreUpgrades store.StoreUpgrades
}

type UpgradeKeepers struct {
	// keepers
	AccountKeeper      authkeeper.AccountKeeper
	BankKeeper         bankkeeper.Keeper
	TransferKeeper     transferkeeper.Keeper
	IcqKeeper          icqkeeper.Keeper
	CronKeeper         cronkeeper.Keeper
	TokenFactoryKeeper *tokenfactorykeeper.Keeper
	FeeBurnerKeeper    *feeburnerkeeper.Keeper
	SlashingKeeper     slashingkeeper.Keeper
	ParamsKeeper       paramskeeper.Keeper
	ContractManager    contractmanagerkeeper.Keeper
	AdminModule        adminmodulekeeper.Keeper
	ConsensusKeeper    *consensuskeeper.Keeper
	MarketmapKeeper    *marketmapkeeper.Keeper
	FeeMarketKeeper    *feemarketkeeper.Keeper
	DynamicfeesKeeper  *dynamicfeeskeeper.Keeper
	StakingKeeper      *stakingkeeper.Keeper
	DexKeeper          *dexkeeper.Keeper
	IbcRateLimitKeeper *ibcratelimitkeeper.Keeper
	ChannelKeeper      *channelkeeper.Keeper
	WasmKeeper         *wasmkeeper.Keeper
	HarpoonKeeper      *harpoonkeeper.Keeper
	RevenueKeeper      *revenuekeeper.Keeper
	// subspaces
	GlobalFeeSubspace paramtypes.Subspace
}

type StoreKeys interface {
	GetKey(string) *store.KVStoreKey
}
