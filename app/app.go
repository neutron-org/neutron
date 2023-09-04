package app

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cosmos/interchain-security/v3/testutil/integration"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	tendermint "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	proposalhandler "github.com/skip-mev/pob/abci"
	"github.com/skip-mev/pob/mempool"
	"github.com/skip-mev/pob/x/builder"
	builderkeeper "github.com/skip-mev/pob/x/builder/keeper"
	rewardsaddressprovider "github.com/skip-mev/pob/x/builder/rewards_address_provider"
	buildertypes "github.com/skip-mev/pob/x/builder/types"

	"github.com/neutron-org/neutron/docs"

	"github.com/neutron-org/neutron/app/upgrades"
	"github.com/neutron-org/neutron/app/upgrades/nextupgrade"
	v044 "github.com/neutron-org/neutron/app/upgrades/v0.4.4"
	v3 "github.com/neutron-org/neutron/app/upgrades/v3"
	"github.com/neutron-org/neutron/x/cron"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	tmos "github.com/cometbft/cometbft/libs/os"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gaia/v11/x/globalfee"
	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v7/modules/core/02-client"
	ibcporttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	"github.com/cosmos/interchain-security/v3/legacy_ibc_testing/core"
	ibctesting "github.com/cosmos/interchain-security/v3/legacy_ibc_testing/testing"
	"github.com/spf13/cast"

	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	cronkeeper "github.com/neutron-org/neutron/x/cron/keeper"
	crontypes "github.com/neutron-org/neutron/x/cron/types"

	"github.com/neutron-org/neutron/x/tokenfactory"
	tokenfactorykeeper "github.com/neutron-org/neutron/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/neutron-org/neutron/x/tokenfactory/types"

	"github.com/cosmos/admin-module/x/adminmodule"
	adminmodulecli "github.com/cosmos/admin-module/x/adminmodule/client/cli"
	adminmodulekeeper "github.com/cosmos/admin-module/x/adminmodule/keeper"
	admintypes "github.com/cosmos/admin-module/x/adminmodule/types"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	appparams "github.com/neutron-org/neutron/app/params"
	"github.com/neutron-org/neutron/wasmbinding"
	"github.com/neutron-org/neutron/x/contractmanager"
	contractmanagermodulekeeper "github.com/neutron-org/neutron/x/contractmanager/keeper"
	contractmanagermoduletypes "github.com/neutron-org/neutron/x/contractmanager/types"
	"github.com/neutron-org/neutron/x/feeburner"
	feeburnerkeeper "github.com/neutron-org/neutron/x/feeburner/keeper"
	feeburnertypes "github.com/neutron-org/neutron/x/feeburner/types"
	"github.com/neutron-org/neutron/x/feerefunder"
	feekeeper "github.com/neutron-org/neutron/x/feerefunder/keeper"
	ibchooks "github.com/neutron-org/neutron/x/ibc-hooks"
	ibchookstypes "github.com/neutron-org/neutron/x/ibc-hooks/types"
	"github.com/neutron-org/neutron/x/interchainqueries"
	interchainqueriesmodulekeeper "github.com/neutron-org/neutron/x/interchainqueries/keeper"
	interchainqueriesmoduletypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	"github.com/neutron-org/neutron/x/interchaintxs"
	interchaintxskeeper "github.com/neutron-org/neutron/x/interchaintxs/keeper"
	interchaintxstypes "github.com/neutron-org/neutron/x/interchaintxs/types"
	transferSudo "github.com/neutron-org/neutron/x/transfer"
	wrapkeeper "github.com/neutron-org/neutron/x/transfer/keeper"

	feetypes "github.com/neutron-org/neutron/x/feerefunder/types"

	ccvconsumer "github.com/cosmos/interchain-security/v3/x/ccv/consumer"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v3/x/ccv/consumer/keeper"
	ccvconsumertypes "github.com/cosmos/interchain-security/v3/x/ccv/consumer/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router"
	routerkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router/keeper"
	routertypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router/types"
)

const (
	Name = "neutrond"
)

var (
	Upgrades = []upgrades.Upgrade{v3.Upgrade, v044.Upgrade, nextupgrade.Upgrade}

	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		ibc.AppModuleBasic{},
		ica.AppModuleBasic{},
		tendermint.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		transferSudo.AppModuleBasic{},
		vesting.AppModuleBasic{},
		ccvconsumer.AppModuleBasic{},
		wasm.AppModuleBasic{},
		tokenfactory.AppModuleBasic{},
		interchainqueries.AppModuleBasic{},
		interchaintxs.AppModuleBasic{},
		feerefunder.AppModuleBasic{},
		feeburner.AppModuleBasic{},
		contractmanager.AppModuleBasic{},
		cron.AppModuleBasic{},
		adminmodule.NewAppModuleBasic(
			govclient.NewProposalHandler(
				adminmodulecli.NewSubmitParamChangeProposalTxCmd,
			),
			govclient.NewProposalHandler(
				adminmodulecli.NewCmdSubmitUpgradeProposal,
			),
			govclient.NewProposalHandler(
				adminmodulecli.NewCmdSubmitCancelUpgradeProposal,
			),
		),
		ibchooks.AppModuleBasic{},
		router.AppModuleBasic{},
		builder.AppModuleBasic{},
		globalfee.AppModule{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:                    nil,
		buildertypes.ModuleName:                       nil,
		ibctransfertypes.ModuleName:                   {authtypes.Minter, authtypes.Burner},
		icatypes.ModuleName:                           nil,
		wasm.ModuleName:                               {authtypes.Burner},
		interchainqueriesmoduletypes.ModuleName:       nil,
		feetypes.ModuleName:                           nil,
		feeburnertypes.ModuleName:                     nil,
		ccvconsumertypes.ConsumerRedistributeName:     {authtypes.Burner},
		ccvconsumertypes.ConsumerToSendToProviderName: nil,
		tokenfactorytypes.ModuleName:                  {authtypes.Minter, authtypes.Burner},
		crontypes.ModuleName:                          nil,
	}
)

var (
	_ runtime.AppI            = (*App)(nil)
	_ servertypes.Application = (*App)(nil)
	_ ibctesting.TestingApp   = (*App)(nil)
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+Name)
}

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp

	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	configurator module.Configurator

	encodingConfig appparams.EncodingConfig

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper     authkeeper.AccountKeeper
	AdminmoduleKeeper adminmodulekeeper.Keeper
	AuthzKeeper       authzkeeper.Keeper
	BankKeeper        bankkeeper.BaseKeeper
	// BuilderKeeper is the keeper that handles processing auction transactions
	BuilderKeeper       builderkeeper.Keeper
	CapabilityKeeper    *capabilitykeeper.Keeper
	SlashingKeeper      slashingkeeper.Keeper
	CrisisKeeper        crisiskeeper.Keeper
	UpgradeKeeper       upgradekeeper.Keeper
	ParamsKeeper        paramskeeper.Keeper
	IBCKeeper           *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	ICAControllerKeeper icacontrollerkeeper.Keeper
	ICAHostKeeper       icahostkeeper.Keeper
	EvidenceKeeper      evidencekeeper.Keeper
	TransferKeeper      wrapkeeper.KeeperTransferWrapper
	FeeGrantKeeper      feegrantkeeper.Keeper
	FeeKeeper           *feekeeper.Keeper
	FeeBurnerKeeper     *feeburnerkeeper.Keeper
	ConsumerKeeper      ccvconsumerkeeper.Keeper
	TokenFactoryKeeper  *tokenfactorykeeper.Keeper
	CronKeeper          cronkeeper.Keeper
	RouterKeeper        *routerkeeper.Keeper

	RouterModule router.AppModule

	HooksTransferIBCModule *ibchooks.IBCMiddleware
	HooksICS4Wrapper       ibchooks.ICS4Middleware

	// make scoped keepers public for test purposes
	ScopedIBCKeeper         capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper    capabilitykeeper.ScopedKeeper
	ScopedWasmKeeper        capabilitykeeper.ScopedKeeper
	ScopedInterTxKeeper     capabilitykeeper.ScopedKeeper
	ScopedCCVConsumerKeeper capabilitykeeper.ScopedKeeper

	InterchainQueriesKeeper interchainqueriesmodulekeeper.Keeper
	InterchainTxsKeeper     interchaintxskeeper.Keeper
	ContractManagerKeeper   contractmanagermodulekeeper.Keeper

	ConsensusParamsKeeper consensusparamkeeper.Keeper

	WasmKeeper wasm.Keeper

	// mm is the module manager
	mm *module.Manager

	// sm is the simulation manager
	sm *module.SimulationManager

	// Custom checkTx handler
	checkTxHandler proposalhandler.CheckTx
}

func (app *App) GetTestBankKeeper() integration.TestBankKeeper {
	return app.BankKeeper
}

func (app *App) GetTestAccountKeeper() integration.TestAccountKeeper {
	return app.AccountKeeper
}

func (app *App) GetTestSlashingKeeper() integration.TestSlashingKeeper {
	return app.SlashingKeeper
}

func (app *App) GetTestEvidenceKeeper() integration.TestEvidenceKeeper {
	return app.EvidenceKeeper
}

// New returns a reference to an initialized blockchain app
func New(
	logger log.Logger,
	chainID string,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig appparams.EncodingConfig,
	appOpts servertypes.AppOptions,
	wasmOpts []wasm.Option,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	appCodec := encodingConfig.Marshaler
	legacyAmino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	config := mempool.NewDefaultAuctionFactory(encodingConfig.TxConfig.TxDecoder())
	// 0 - unlimited amount of txs
	maxTx := 0
	mempool := mempool.NewAuctionMempool(encodingConfig.TxConfig.TxDecoder(), encodingConfig.TxConfig.TxEncoder(), maxTx, config)

	bApp := baseapp.NewBaseApp(Name, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetMempool(mempool)

	keys := sdk.NewKVStoreKeys(
		authzkeeper.StoreKey, authtypes.StoreKey, banktypes.StoreKey, slashingtypes.StoreKey,
		paramstypes.StoreKey, ibchost.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey, icacontrollertypes.StoreKey,
		icahosttypes.StoreKey, capabilitytypes.StoreKey,
		interchainqueriesmoduletypes.StoreKey, contractmanagermoduletypes.StoreKey, interchaintxstypes.StoreKey, wasm.StoreKey, feetypes.StoreKey,
		feeburnertypes.StoreKey, admintypes.StoreKey, ccvconsumertypes.StoreKey, tokenfactorytypes.StoreKey, routertypes.StoreKey,
		crontypes.StoreKey, ibchookstypes.StoreKey, consensusparamtypes.StoreKey, crisistypes.StoreKey, buildertypes.StoreKey,
	)
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey, feetypes.MemStoreKey)

	app := &App{
		BaseApp:           bApp,
		cdc:               legacyAmino,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
		encodingConfig:    encodingConfig,
	}

	app.ParamsKeeper = initParamsKeeper(appCodec, legacyAmino, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(appCodec, keys[consensusparamtypes.StoreKey], authtypes.NewModuleAddress(admintypes.ModuleName).String())
	bApp.SetParamStore(&app.ConsensusParamsKeeper)

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])

	// grant capabilities for the ibc and ibc-transfer modules
	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	scopedICAControllerKeeper := app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedWasmKeeper := app.CapabilityKeeper.ScopeToModule(wasm.ModuleName)
	scopedInterTxKeeper := app.CapabilityKeeper.ScopeToModule(interchaintxstypes.ModuleName)
	scopedCCVConsumerKeeper := app.CapabilityKeeper.ScopeToModule(ccvconsumertypes.ModuleName)

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		maccPerms,
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authtypes.NewModuleAddress(admintypes.ModuleName).String(),
	)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		keys[authz.ModuleName], appCodec, app.MsgServiceRouter(), app.AccountKeeper,
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		app.AccountKeeper,
		app.BlockedAddrs(),
		authtypes.NewModuleAddress(admintypes.ModuleName).String(),
	)

	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		keys[slashingtypes.StoreKey],
		&app.ConsumerKeeper,
		authtypes.NewModuleAddress(admintypes.ModuleName).String(),
	)
	app.CrisisKeeper = *crisiskeeper.NewKeeper(
		appCodec,
		keys[crisistypes.StoreKey],
		invCheckPeriod,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(admintypes.ModuleName).String(),
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, keys[feegrant.StoreKey], app.AccountKeeper)
	app.UpgradeKeeper = *upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		app.BaseApp,
		authtypes.NewModuleAddress(admintypes.ModuleName).String(),
	)

	// ... other modules keepers
	// pre-initialize ConsumerKeeper to satsfy ibckeeper.NewKeeper
	// which would panic on nil or zero keeper
	// ConsumerKeeper implements StakingKeeper but all function calls result in no-ops so this is safe
	// communication over IBC is not affected by these changes
	app.ConsumerKeeper = ccvconsumerkeeper.NewNonZeroKeeper(
		appCodec,
		keys[ccvconsumertypes.StoreKey],
		app.GetSubspace(ccvconsumertypes.ModuleName),
	)

	// Create IBC Keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec, keys[ibchost.StoreKey], app.GetSubspace(ibchost.ModuleName), &app.ConsumerKeeper, app.UpgradeKeeper, scopedIBCKeeper,
	)

	app.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		appCodec, keys[icacontrollertypes.StoreKey], app.GetSubspace(icacontrollertypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper, // may be replaced with middleware such as ics29 feerefunder
		app.IBCKeeper.ChannelKeeper, &app.IBCKeeper.PortKeeper,
		scopedICAControllerKeeper, app.MsgServiceRouter(),
	)

	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec, keys[icahosttypes.StoreKey], app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper, // may be replaced with middleware such as ics29 feerefunder
		app.IBCKeeper.ChannelKeeper, &app.IBCKeeper.PortKeeper,
		app.AccountKeeper, scopedICAHostKeeper, app.MsgServiceRouter(),
	)

	app.ContractManagerKeeper = *contractmanagermodulekeeper.NewKeeper(
		appCodec,
		keys[contractmanagermoduletypes.StoreKey],
		keys[contractmanagermoduletypes.MemStoreKey],
		&app.WasmKeeper,
	)

	app.FeeKeeper = feekeeper.NewKeeper(appCodec, keys[feetypes.StoreKey], memKeys[feetypes.MemStoreKey], app.IBCKeeper.ChannelKeeper, app.BankKeeper)
	feeModule := feerefunder.NewAppModule(appCodec, *app.FeeKeeper, app.AccountKeeper, app.BankKeeper)

	app.FeeBurnerKeeper = feeburnerkeeper.NewKeeper(
		appCodec,
		keys[feeburnertypes.StoreKey],
		keys[feeburnertypes.MemStoreKey],
		app.AccountKeeper,
		app.BankKeeper,
	)
	feeBurnerModule := feeburner.NewAppModule(appCodec, *app.FeeBurnerKeeper)

	// RouterKeeper must be created before TransferKeeper
	app.RouterKeeper = routerkeeper.NewKeeper(
		appCodec,
		app.keys[routertypes.StoreKey],
		app.GetSubspace(routertypes.ModuleName),
		app.TransferKeeper.Keeper,
		app.IBCKeeper.ChannelKeeper,
		app.FeeBurnerKeeper,
		app.BankKeeper,
		app.IBCKeeper.ChannelKeeper,
	)
	wasmHooks := ibchooks.NewWasmHooks(nil, sdk.GetConfig().GetBech32AccountAddrPrefix()) // The contract keeper needs to be set later
	app.HooksICS4Wrapper = ibchooks.NewICS4Middleware(
		app.IBCKeeper.ChannelKeeper,
		app.RouterKeeper,
		&wasmHooks,
	)

	// Create Transfer Keepers
	app.TransferKeeper = wrapkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.HooksICS4Wrapper, // essentially still app.IBCKeeper.ChannelKeeper under the hood because no hook overrides
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		scopedTransferKeeper,
		app.FeeKeeper,
		app.ContractManagerKeeper,
	)

	app.RouterKeeper.SetTransferKeeper(app.TransferKeeper.Keeper)

	transferModule := transferSudo.NewAppModule(app.TransferKeeper)

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec, keys[evidencetypes.StoreKey], &app.ConsumerKeeper, app.SlashingKeeper,
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	app.ConsumerKeeper = ccvconsumerkeeper.NewKeeper(
		appCodec,
		keys[ccvconsumertypes.StoreKey],
		app.GetSubspace(ccvconsumertypes.ModuleName),
		scopedCCVConsumerKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.IBCKeeper.ConnectionKeeper,
		app.IBCKeeper.ClientKeeper,
		app.SlashingKeeper,
		app.BankKeeper,
		app.AccountKeeper,
		app.TransferKeeper.Keeper, // we cant use our transfer wrapper type here because of interface incompatibility, it looks safe to use underlying transfer keeper.
		// Since the keeper is only used to send reward to provider chain
		app.IBCKeeper,
		authtypes.FeeCollectorName,
	)
	app.ConsumerKeeper = *app.ConsumerKeeper.SetHooks(app.SlashingKeeper.Hooks())
	consumerModule := ccvconsumer.NewAppModule(app.ConsumerKeeper, app.GetSubspace(ccvconsumertypes.ModuleName))

	tokenFactoryKeeper := tokenfactorykeeper.NewKeeper(
		appCodec,
		app.keys[tokenfactorytypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper.WithMintCoinsRestriction(tokenfactorytypes.NewTokenFactoryDenomMintCoinsRestriction()),
	)
	app.TokenFactoryKeeper = &tokenFactoryKeeper

	app.BuilderKeeper = builderkeeper.NewKeeperWithRewardsAddressProvider(
		appCodec,
		keys[buildertypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		// 25% of rewards should be sent to the redistribute address
		rewardsaddressprovider.NewFixedAddressRewardsAddressProvider(app.AccountKeeper.GetModuleAddress(ccvconsumertypes.ConsumerRedistributeName)),
		authtypes.NewModuleAddress(admintypes.ModuleName).String(),
	)

	wasmDir := filepath.Join(homePath, "wasm")
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}

	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	// NOTE: we need staking feature here even if there is no staking module anymore because cosmwasm-std in the CosmWasm SDK requires this feature
	// NOTE: cosmwasm_1_2 feature enables GovMsg::VoteWeighted, which doesn't work with Neutron, because it uses its own custom governance,
	//       however, cosmwasm_1_2 also enables WasmMsg::Instantiate2, which works as one could expect
	supportedFeatures := "iterator,stargate,staking,neutron,cosmwasm_1_1,cosmwasm_1_2"

	// register the proposal types
	adminRouterLegacy := govv1beta1.NewRouter()
	adminRouterLegacy.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(&app.UpgradeKeeper)).
		AddRoute(ibchost.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper))

	app.AdminmoduleKeeper = *adminmodulekeeper.NewKeeper(
		appCodec,
		keys[admintypes.StoreKey],
		keys[admintypes.MemStoreKey],
		adminRouterLegacy,
		app.MsgServiceRouter(),
		IsConsumerProposalAllowlisted,
		isSdkMessageWhitelisted,
	)
	adminModule := adminmodule.NewAppModule(appCodec, app.AdminmoduleKeeper)

	app.InterchainQueriesKeeper = *interchainqueriesmodulekeeper.NewKeeper(
		appCodec,
		keys[interchainqueriesmoduletypes.StoreKey],
		keys[interchainqueriesmoduletypes.MemStoreKey],
		app.IBCKeeper,
		app.BankKeeper,
		app.ContractManagerKeeper,
		interchainqueriesmodulekeeper.Verifier{},
		interchainqueriesmodulekeeper.TransactionVerifier{},
	)
	app.InterchainTxsKeeper = *interchaintxskeeper.NewKeeper(
		appCodec,
		keys[interchaintxstypes.StoreKey],
		memKeys[interchaintxstypes.MemStoreKey],
		app.IBCKeeper.ChannelKeeper,
		app.ICAControllerKeeper,
		app.ContractManagerKeeper,
		app.FeeKeeper,
	)

	app.CronKeeper = *cronkeeper.NewKeeper(appCodec, keys[crontypes.StoreKey], keys[crontypes.MemStoreKey], app.AccountKeeper)
	wasmOpts = append(wasmbinding.RegisterCustomPlugins(&app.InterchainTxsKeeper, &app.InterchainQueriesKeeper, app.TransferKeeper, &app.AdminmoduleKeeper, app.FeeBurnerKeeper, app.FeeKeeper, &app.BankKeeper, app.TokenFactoryKeeper, &app.CronKeeper), wasmOpts...)

	queryPlugins := wasmkeeper.WithQueryPlugins(
		&wasmkeeper.QueryPlugins{Stargate: wasmkeeper.AcceptListStargateQuerier(wasmbinding.AcceptedStargateQueries(), app.GRPCQueryRouter(), appCodec)})
	wasmOpts = append(wasmOpts, queryPlugins)

	app.WasmKeeper = wasm.NewKeeper(
		appCodec,
		keys[wasm.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		nil,
		nil,
		app.IBCKeeper.ChannelKeeper, // may be replaced with middleware such as ics29 feerefunder
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		scopedWasmKeeper,
		app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		supportedFeatures,
		authtypes.NewModuleAddress(admintypes.ModuleName).String(),
		wasmOpts...,
	)
	wasmHooks.ContractKeeper = &app.WasmKeeper

	app.CronKeeper.WasmMsgServer = wasmkeeper.NewMsgServerImpl(&app.WasmKeeper)
	cronModule := cron.NewAppModule(appCodec, &app.CronKeeper)

	transferIBCModule := transferSudo.NewIBCModule(app.TransferKeeper)
	// receive call order: wasmHooks#OnRecvPacketOverride(transferIbcModule#OnRecvPacket())
	ibcHooksMiddleware := ibchooks.NewIBCMiddleware(&transferIBCModule, &app.HooksICS4Wrapper)
	app.HooksTransferIBCModule = &ibcHooksMiddleware

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := ibcporttypes.NewRouter()

	icaModule := ica.NewAppModule(&app.ICAControllerKeeper, &app.ICAHostKeeper)

	var icaControllerStack ibcporttypes.IBCModule

	icaControllerStack = interchaintxs.NewIBCModule(app.InterchainTxsKeeper)
	icaControllerStack = icacontroller.NewIBCMiddleware(icaControllerStack, app.ICAControllerKeeper)

	icaHostIBCModule := icahost.NewIBCModule(app.ICAHostKeeper)

	interchainQueriesModule := interchainqueries.NewAppModule(appCodec, app.InterchainQueriesKeeper, app.AccountKeeper, app.BankKeeper)
	interchainTxsModule := interchaintxs.NewAppModule(appCodec, app.InterchainTxsKeeper, app.AccountKeeper, app.BankKeeper)
	contractManagerModule := contractmanager.NewAppModule(appCodec, app.ContractManagerKeeper)
	ibcHooksModule := ibchooks.NewAppModule(app.AccountKeeper)

	app.RouterModule = router.NewAppModule(app.RouterKeeper)

	ibcStack := router.NewIBCMiddleware(
		app.HooksTransferIBCModule,
		app.RouterKeeper,
		0,
		routerkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
		routerkeeper.DefaultRefundTransferPacketTimeoutTimestamp,
	)

	ibcRouter.AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(icahosttypes.SubModuleName, icaHostIBCModule).
		AddRoute(ibctransfertypes.ModuleName, ibcStack).
		AddRoute(interchaintxstypes.ModuleName, icaControllerStack).
		AddRoute(wasm.ModuleName, wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper, app.IBCKeeper.ChannelKeeper)).
		AddRoute(ccvconsumertypes.ModuleName, consumerModule)
	app.IBCKeeper.SetRouter(ibcRouter)

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.setupUpgradeStoreLoaders()

	app.mm = module.NewManager(
		auth.NewAppModule(appCodec, app.AccountKeeper, nil, app.GetSubspace(authtypes.ModuleName)),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.ConsumerKeeper, app.GetSubspace(slashingtypes.ModuleName)),
		upgrade.NewAppModule(&app.UpgradeKeeper),
		wasm.NewAppModule(appCodec, &app.WasmKeeper, app.AccountKeeper, app.BankKeeper, app.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
		evidence.NewAppModule(app.EvidenceKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		params.NewAppModule(app.ParamsKeeper),
		transferModule,
		consumerModule,
		icaModule,
		app.RouterModule,
		interchainQueriesModule,
		interchainTxsModule,
		feeModule,
		feeBurnerModule,
		contractManagerModule,
		adminModule,
		ibcHooksModule,
		tokenfactory.NewAppModule(appCodec, *app.TokenFactoryKeeper, app.AccountKeeper, app.BankKeeper),
		cronModule,
		globalfee.NewAppModule(app.GetSubspace(globalfee.ModuleName)),
		builder.NewAppModule(appCodec, app.BuilderKeeper),
		crisis.NewAppModule(&app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)), // always be last to make sure that it checks for all invariants and not only part of them
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator feerefunder pool to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		buildertypes.ModuleName,
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		vestingtypes.ModuleName,
		ibchost.ModuleName,
		ibctransfertypes.ModuleName,
		authtypes.ModuleName,
		authz.ModuleName,
		banktypes.ModuleName,
		crisistypes.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		ccvconsumertypes.ModuleName,
		tokenfactorytypes.ModuleName,
		icatypes.ModuleName,
		interchainqueriesmoduletypes.ModuleName,
		interchaintxstypes.ModuleName,
		contractmanagermoduletypes.ModuleName,
		wasm.ModuleName,
		feetypes.ModuleName,
		feeburnertypes.ModuleName,
		admintypes.ModuleName,
		ibchookstypes.ModuleName,
		routertypes.ModuleName,
		crontypes.ModuleName,
		globalfee.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
		buildertypes.ModuleName,
		crisistypes.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		authz.ModuleName,
		banktypes.ModuleName,
		slashingtypes.ModuleName,
		vestingtypes.ModuleName,
		evidencetypes.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibchost.ModuleName,
		ibctransfertypes.ModuleName,
		ccvconsumertypes.ModuleName,
		tokenfactorytypes.ModuleName,
		icatypes.ModuleName,
		interchainqueriesmoduletypes.ModuleName,
		interchaintxstypes.ModuleName,
		contractmanagermoduletypes.ModuleName,
		wasm.ModuleName,
		feetypes.ModuleName,
		feeburnertypes.ModuleName,
		admintypes.ModuleName,
		ibchookstypes.ModuleName,
		routertypes.ModuleName,
		crontypes.ModuleName,
		globalfee.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(
		buildertypes.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		ibctransfertypes.ModuleName,
		authz.ModuleName,
		banktypes.ModuleName,
		vestingtypes.ModuleName,
		slashingtypes.ModuleName,
		crisistypes.ModuleName,
		ibchost.ModuleName,
		evidencetypes.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		feegrant.ModuleName,
		ccvconsumertypes.ModuleName,
		tokenfactorytypes.ModuleName,
		icatypes.ModuleName,
		interchainqueriesmoduletypes.ModuleName,
		interchaintxstypes.ModuleName,
		contractmanagermoduletypes.ModuleName,
		wasm.ModuleName,
		feetypes.ModuleName,
		feeburnertypes.ModuleName,
		admintypes.ModuleName,
		ibchookstypes.ModuleName, // after auth keeper
		routertypes.ModuleName,
		crontypes.ModuleName,
		globalfee.ModuleName,
	)

	app.mm.RegisterInvariants(&app.CrisisKeeper)
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	app.setupUpgradeHandlers()

	// create the simulation manager and define the order of the modules for deterministic simulations
	app.sm = module.NewSimulationManager(
		auth.NewAppModule(appCodec, app.AccountKeeper, nil, app.GetSubspace(authtypes.ModuleName)),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, nil, app.GetSubspace(slashingtypes.ModuleName)),
		wasm.NewAppModule(appCodec, &app.WasmKeeper, app.AccountKeeper, app.BankKeeper, app.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
		evidence.NewAppModule(app.EvidenceKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		params.NewAppModule(app.ParamsKeeper),
		transferModule,
		consumerModule,
		icaModule,
		app.RouterModule,
		interchainQueriesModule,
		interchainTxsModule,
		feeBurnerModule,
		cronModule,
	)
	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)

	anteHandler, err := NewAnteHandler(
		HandlerOptions{
			HandlerOptions: ante.HandlerOptions{
				AccountKeeper:   app.AccountKeeper,
				BankKeeper:      app.BankKeeper,
				FeegrantKeeper:  app.FeeGrantKeeper,
				SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
				SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			},
			IBCKeeper:         app.IBCKeeper,
			WasmConfig:        &wasmConfig,
			TXCounterStoreKey: keys[wasm.StoreKey],
			ConsumerKeeper:    app.ConsumerKeeper,
			buildKeeper:       app.BuilderKeeper,
			mempool:           mempool,
			GlobalFeeSubspace: app.GetSubspace(globalfee.ModuleName),
		},
		app.Logger(),
	)
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)
	app.SetEndBlocker(app.EndBlocker)

	handler := proposalhandler.NewProposalHandler(
		mempool,
		bApp.Logger(),
		anteHandler,
		encodingConfig.TxConfig.TxEncoder(),
		encodingConfig.TxConfig.TxDecoder(),
	)
	app.SetPrepareProposal(handler.PrepareProposalHandler())
	app.SetProcessProposal(handler.ProcessProposalHandler())

	checkTxHandler := proposalhandler.NewCheckTxHandler(
		app.BaseApp,
		encodingConfig.TxConfig.TxDecoder(),
		mempool,
		anteHandler,
		chainID,
	)
	app.SetCheckTx(checkTxHandler.CheckTx())

	// must be before Loading version
	// requires the snapshot store to be created and registered as a BaseAppOption
	// see cmd/wasmd/root.go: 206 - 214 approx
	if manager := app.SnapshotManager(); manager != nil {
		err := manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.WasmKeeper),
		)
		if err != nil {
			panic(fmt.Errorf("failed to register snapshot extension: %s", err))
		}
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}

		ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})

		// Initialize pinned codes in wasmvm as they are not persisted there
		if err := app.WasmKeeper.InitializePinnedCodes(ctx); err != nil {
			tmos.Exit(fmt.Sprintf("failed initialize pinned codes %s", err))
		}
	}

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper
	app.ScopedWasmKeeper = scopedWasmKeeper
	app.ScopedInterTxKeeper = scopedInterTxKeeper
	app.ScopedCCVConsumerKeeper = scopedCCVConsumerKeeper

	return app
}

func (app *App) setupUpgradeStoreLoaders() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	for _, upgrade := range Upgrades {
		if upgradeInfo.Name == upgrade.UpgradeName {
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &upgrade.StoreUpgrades))
		}
	}
}

func (app *App) setupUpgradeHandlers() {
	for _, upgrade := range Upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgrade.UpgradeName,
			upgrade.CreateUpgradeHandler(
				app.mm,
				app.configurator,
				&upgrades.UpgradeKeepers{
					FeeBurnerKeeper:    app.FeeBurnerKeeper,
					CronKeeper:         app.CronKeeper,
					IcqKeeper:          app.InterchainQueriesKeeper,
					TokenFactoryKeeper: app.TokenFactoryKeeper,
					SlashingKeeper:     app.SlashingKeeper,
					ParamsKeeper:       app.ParamsKeeper,
					CapabilityKeeper:   app.CapabilityKeeper,
					GlobalFeeSubspace:  app.GetSubspace(globalfee.ModuleName),
				},
				app,
				app.AppCodec(),
			),
		)
	}
}

// CheckTx will check the transaction with the provided checkTxHandler. We override the default
// handler so that we can verify bid transactions before they are inserted into the mempool.
// With the POB CheckTx, we can verify the bid transaction and all of the bundled transactions
// before inserting the bid transaction into the mempool.
func (app *App) CheckTx(req abci.RequestCheckTx) abci.ResponseCheckTx {
	return app.checkTxHandler(req)
}

// SetCheckTx sets the checkTxHandler for the app.
func (app *App) SetCheckTx(handler proposalhandler.CheckTx) {
	app.checkTxHandler = handler
}

// Name returns the name of the App
func (app *App) Name() string { return app.BaseApp.Name() }

// GetBaseApp returns the base app of the application
func (app *App) GetBaseApp() *baseapp.BaseApp { return app.BaseApp }

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *App) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// BlockedAddrs returns the set of addresses that are not allowed
// to send and receive funds
func (app *App) BlockedAddrs() map[string]bool {
	// Remove the fee-pool from the group of blocked recipient addresses in bank
	// this is required for the consumer chain to be able to send tokens to
	// the provider chain
	bankBlockedAddrs := app.ModuleAccountAddrs()
	delete(bankBlockedAddrs, authtypes.NewModuleAddress(
		ccvconsumertypes.ConsumerToSendToProviderName).String())

	return bankBlockedAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.cdc
}

// AppCodec returns an app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *App) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register app's swagger ui
	if apiConfig.Swagger {
		app.RegisterSwaggerUI(apiSvr)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(clientCtx, app.BaseApp.GRPCQueryRouter(), app.interfaceRegistry, app.Query)
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibchost.ModuleName)
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)

	paramsKeeper.Subspace(routertypes.ModuleName).WithKeyTable(routertypes.ParamKeyTable())

	paramsKeeper.Subspace(ccvconsumertypes.ModuleName)
	paramsKeeper.Subspace(tokenfactorytypes.ModuleName)

	paramsKeeper.Subspace(interchainqueriesmoduletypes.ModuleName)
	paramsKeeper.Subspace(interchaintxstypes.ModuleName)
	paramsKeeper.Subspace(wasm.ModuleName)
	paramsKeeper.Subspace(feetypes.ModuleName)
	paramsKeeper.Subspace(feeburnertypes.ModuleName)
	paramsKeeper.Subspace(crontypes.ModuleName)
	paramsKeeper.Subspace(globalfee.ModuleName)

	return paramsKeeper
}

// SimulationManager implements the SimulationApp interface
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

func (app *App) RegisterSwaggerUI(apiSvr *api.Server) {
	staticSubDir, err := fs.Sub(docs.Docs, "static")
	if err != nil {
		app.Logger().Error(fmt.Sprintf("failed to register swagger-ui route: %s", err))
		return
	}

	staticServer := http.FileServer(http.FS(staticSubDir))
	apiSvr.Router.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

// ConsumerApp interface implementations for e2e tests

// GetTxConfig implements the TestingApp interface.
func (app *App) GetTxConfig() client.TxConfig {
	return app.encodingConfig.TxConfig
}

// GetIBCKeeper implements the TestingApp interface.
func (app *App) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

// GetStakingKeeper implements the TestingApp interface.
func (app *App) GetStakingKeeper() core.StakingKeeper {
	return app.ConsumerKeeper
}

// GetScopedIBCKeeper implements the TestingApp interface.
func (app *App) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

// GetConsumerKeeper implements the ConsumerApp interface.
func (app *App) GetConsumerKeeper() ccvconsumerkeeper.Keeper {
	return app.ConsumerKeeper
}

func (app *App) RegisterNodeService(clientCtx client.Context) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}
