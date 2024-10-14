package upgrades

import (
	store "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	adminmodulekeeper "github.com/cosmos/admin-module/v2/x/adminmodule/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v5/x/ccv/consumer/keeper"
	auctionkeeper "github.com/skip-mev/block-sdk/v2/x/auction/keeper"
	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"
	marketmapkeeper "github.com/skip-mev/slinky/x/marketmap/keeper"

	dexkeeper "github.com/neutron-org/neutron/v5/x/dex/keeper"
	ibcratelimitkeeper "github.com/neutron-org/neutron/v5/x/ibc-rate-limit/keeper"

	dynamicfeeskeeper "github.com/neutron-org/neutron/v5/x/dynamicfees/keeper"

	contractmanagerkeeper "github.com/neutron-org/neutron/v5/x/contractmanager/keeper"
	cronkeeper "github.com/neutron-org/neutron/v5/x/cron/keeper"
	feeburnerkeeper "github.com/neutron-org/neutron/v5/x/feeburner/keeper"
	icqkeeper "github.com/neutron-org/neutron/v5/x/interchainqueries/keeper"
	tokenfactorykeeper "github.com/neutron-org/neutron/v5/x/tokenfactory/keeper"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	transferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
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
	CapabilityKeeper   *capabilitykeeper.Keeper
	AuctionKeeper      auctionkeeper.Keeper
	ContractManager    contractmanagerkeeper.Keeper
	AdminModule        adminmodulekeeper.Keeper
	ConsensusKeeper    *consensuskeeper.Keeper
	ConsumerKeeper     *ccvconsumerkeeper.Keeper
	MarketmapKeeper    *marketmapkeeper.Keeper
	FeeMarketKeeper    *feemarketkeeper.Keeper
	DynamicfeesKeeper  *dynamicfeeskeeper.Keeper
	DexKeeper          *dexkeeper.Keeper
	IbcRateLimitKeeper *ibcratelimitkeeper.Keeper
	// subspaces
	GlobalFeeSubspace   paramtypes.Subspace
	CcvConsumerSubspace paramtypes.Subspace
}

type StoreKeys interface {
	GetKey(string) *store.KVStoreKey
}
