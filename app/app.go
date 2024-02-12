package app

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/testutil/sims"
	globalfeetypes "github.com/cosmos/gaia/v11/x/globalfee/types"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward"
	ibctestingtypes "github.com/cosmos/ibc-go/v7/testing/types"
	"github.com/cosmos/interchain-security/v4/testutil/integration"
	ccv "github.com/cosmos/interchain-security/v4/x/ccv/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	tendermint "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	"github.com/neutron-org/neutron/v2/app/upgrades/nextupgrade"
	"github.com/neutron-org/neutron/v2/docs"

	"github.com/neutron-org/neutron/v2/app/upgrades"
	v030 "github.com/neutron-org/neutron/v2/app/upgrades/v0.3.0"
	v044 "github.com/neutron-org/neutron/v2/app/upgrades/v0.4.4"
	v200 "github.com/neutron-org/neutron/v2/app/upgrades/v2.0.0"
	v202 "github.com/neutron-org/neutron/v2/app/upgrades/v2.0.2"

	"github.com/neutron-org/neutron/v2/x/cron"

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
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcporttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/spf13/cast"

	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	cronkeeper "github.com/neutron-org/neutron/v2/x/cron/keeper"
	crontypes "github.com/neutron-org/neutron/v2/x/cron/types"

	"github.com/neutron-org/neutron/v2/x/tokenfactory"
	tokenfactorykeeper "github.com/neutron-org/neutron/v2/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/neutron-org/neutron/v2/x/tokenfactory/types"

	"github.com/cosmos/admin-module/x/adminmodule"
	adminmodulecli "github.com/cosmos/admin-module/x/adminmodule/client/cli"
	adminmodulekeeper "github.com/cosmos/admin-module/x/adminmodule/keeper"
	adminmoduletypes "github.com/cosmos/admin-module/x/adminmodule/types"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	appparams "github.com/neutron-org/neutron/v2/app/params"
	"github.com/neutron-org/neutron/v2/wasmbinding"
	"github.com/neutron-org/neutron/v2/x/contractmanager"
	contractmanagermodulekeeper "github.com/neutron-org/neutron/v2/x/contractmanager/keeper"
	contractmanagermoduletypes "github.com/neutron-org/neutron/v2/x/contractmanager/types"
	"github.com/neutron-org/neutron/v2/x/feeburner"
	feeburnerkeeper "github.com/neutron-org/neutron/v2/x/feeburner/keeper"
	feeburnertypes "github.com/neutron-org/neutron/v2/x/feeburner/types"
	"github.com/neutron-org/neutron/v2/x/feerefunder"
	feekeeper "github.com/neutron-org/neutron/v2/x/feerefunder/keeper"
	ibchooks "github.com/neutron-org/neutron/v2/x/ibc-hooks"
	ibchookstypes "github.com/neutron-org/neutron/v2/x/ibc-hooks/types"
	"github.com/neutron-org/neutron/v2/x/interchainqueries"
	interchainqueriesmodulekeeper "github.com/neutron-org/neutron/v2/x/interchainqueries/keeper"
	interchainqueriesmoduletypes "github.com/neutron-org/neutron/v2/x/interchainqueries/types"
	"github.com/neutron-org/neutron/v2/x/interchaintxs"
	interchaintxskeeper "github.com/neutron-org/neutron/v2/x/interchaintxs/keeper"
	interchaintxstypes "github.com/neutron-org/neutron/v2/x/interchaintxs/types"
	transferSudo "github.com/neutron-org/neutron/v2/x/transfer"
	wrapkeeper "github.com/neutron-org/neutron/v2/x/transfer/keeper"

	feetypes "github.com/neutron-org/neutron/v2/x/feerefunder/types"

	ccvconsumer "github.com/cosmos/interchain-security/v4/x/ccv/consumer"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v4/x/ccv/consumer/keeper"
	ccvconsumertypes "github.com/cosmos/interchain-security/v4/x/ccv/consumer/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	pfmkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/keeper"
	pfmtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"

	"github.com/neutron-org/neutron/v2/x/dex"
	dexkeeper "github.com/neutron-org/neutron/v2/x/dex/keeper"
	dextypes "github.com/neutron-org/neutron/v2/x/dex/types"

	"github.com/neutron-org/neutron/v2/x/ibcswap"
	ibcswapkeeper "github.com/neutron-org/neutron/v2/x/ibcswap/keeper"
	ibcswaptypes "github.com/neutron-org/neutron/v2/x/ibcswap/types"

	gmpmiddleware "github.com/neutron-org/neutron/v2/x/gmp"
)

const (
	Name = "neutrond"
)

var (
	Upgrades = []upgrades.Upgrade{v030.Upgrade, v044.Upgrade, v200.Upgrade, v202.Upgrade, nextupgrade.Upgrade}

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
		packetforward.AppModuleBasic{},
		globalfee.AppModule{},
		dex.AppModuleBasic{},
		ibcswap.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:                    nil,
		ibctransfertypes.ModuleName:                   {authtypes.Minter, authtypes.Burner},
		icatypes.ModuleName:                           nil,
		wasmtypes.ModuleName:                          {},
		interchainqueriesmoduletypes.ModuleName:       nil,
		feetypes.ModuleName:                           nil,
		feeburnertypes.ModuleName:                     nil,
		ccvconsumertypes.ConsumerRedistributeName:     {authtypes.Burner},
		ccvconsumertypes.ConsumerToSendToProviderName: nil,
		tokenfactorytypes.ModuleName:                  {authtypes.Minter, authtypes.Burner},
		crontypes.ModuleName:                          nil,
		dextypes.ModuleName:                           {authtypes.Minter, authtypes.Burner},
		ibcswaptypes.ModuleName:                       {authtypes.Burner},
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
	AccountKeeper       authkeeper.AccountKeeper
	AdminmoduleKeeper   adminmodulekeeper.Keeper
	AuthzKeeper         authzkeeper.Keeper
	BankKeeper          bankkeeper.BaseKeeper
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
	PFMKeeper           *pfmkeeper.Keeper
	DexKeeper           dexkeeper.Keeper
	SwapKeeper          ibcswapkeeper.Keeper

	PFMModule packetforward.AppModule

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

	WasmKeeper wasmkeeper.Keeper

	// mm is the module manager
	mm *module.Manager

	// sm is the simulation manager
	sm *module.SimulationManager
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
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig appparams.EncodingConfig,
	appOpts servertypes.AppOptions,
	wasmOpts []wasmkeeper.Option,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	overrideWasmVariables()

	appCodec := encodingConfig.Marshaler
	legacyAmino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(Name, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := sdk.NewKVStoreKeys(
		authzkeeper.StoreKey, authtypes.StoreKey, banktypes.StoreKey, slashingtypes.StoreKey,
		paramstypes.StoreKey, ibchost.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey, icacontrollertypes.StoreKey,
		icahosttypes.StoreKey, capabilitytypes.StoreKey,
		interchainqueriesmoduletypes.StoreKey, contractmanagermoduletypes.StoreKey, interchaintxstypes.StoreKey, wasmtypes.StoreKey, feetypes.StoreKey,
		feeburnertypes.StoreKey, adminmoduletypes.StoreKey, ccvconsumertypes.StoreKey, tokenfactorytypes.StoreKey, pfmtypes.StoreKey,
		crontypes.StoreKey, ibchookstypes.StoreKey, consensusparamtypes.StoreKey, crisistypes.StoreKey, dextypes.StoreKey,
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
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(appCodec, keys[consensusparamtypes.StoreKey], authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String())
	bApp.SetParamStore(&app.ConsensusParamsKeeper)

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])

	// grant capabilities for the ibc and ibc-transfer modules
	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	scopedICAControllerKeeper := app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedWasmKeeper := app.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)
	scopedInterTxKeeper := app.CapabilityKeeper.ScopeToModule(interchaintxstypes.ModuleName)
	scopedCCVConsumerKeeper := app.CapabilityKeeper.ScopeToModule(ccvconsumertypes.ModuleName)

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		maccPerms,
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		keys[authz.ModuleName], appCodec, app.MsgServiceRouter(), app.AccountKeeper,
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		app.AccountKeeper,
		app.BlockedAddrs(),
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		keys[slashingtypes.StoreKey],
		&app.ConsumerKeeper,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)
	app.CrisisKeeper = *crisiskeeper.NewKeeper(
		appCodec,
		keys[crisistypes.StoreKey],
		invCheckPeriod,
		&app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, keys[feegrant.StoreKey], app.AccountKeeper)
	app.UpgradeKeeper = *upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		app.BaseApp,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
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
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	app.FeeKeeper = feekeeper.NewKeeper(
		appCodec,
		keys[feetypes.StoreKey],
		memKeys[feetypes.MemStoreKey],
		app.IBCKeeper.ChannelKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)
	feeModule := feerefunder.NewAppModule(appCodec, *app.FeeKeeper, app.AccountKeeper, app.BankKeeper)

	app.FeeBurnerKeeper = feeburnerkeeper.NewKeeper(
		appCodec,
		keys[feeburnertypes.StoreKey],
		keys[feeburnertypes.MemStoreKey],
		app.AccountKeeper,
		&app.BankKeeper,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)
	feeBurnerModule := feeburner.NewAppModule(appCodec, *app.FeeBurnerKeeper)

	// PFMKeeper must be created before TransferKeeper
	app.PFMKeeper = pfmkeeper.NewKeeper(
		appCodec,
		app.keys[pfmtypes.StoreKey],
		app.TransferKeeper.Keeper,
		app.IBCKeeper.ChannelKeeper,
		app.FeeBurnerKeeper,
		&app.BankKeeper,
		app.IBCKeeper.ChannelKeeper,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)
	wasmHooks := ibchooks.NewWasmHooks(nil, sdk.GetConfig().GetBech32AccountAddrPrefix()) // The contract keeper needs to be set later
	app.HooksICS4Wrapper = ibchooks.NewICS4Middleware(
		app.IBCKeeper.ChannelKeeper,
		app.PFMKeeper,
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
		&app.BankKeeper,
		scopedTransferKeeper,
		app.FeeKeeper,
		contractmanager.NewSudoLimitWrapper(app.ContractManagerKeeper, &app.WasmKeeper),
	)

	app.PFMKeeper.SetTransferKeeper(app.TransferKeeper.Keeper)

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
		&app.BankKeeper,
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
		&app.WasmKeeper,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)
	app.TokenFactoryKeeper = &tokenFactoryKeeper

	app.BankKeeper.BaseSendKeeper = *app.BankKeeper.BaseSendKeeper.SetHooks(
		banktypes.NewMultiBankHooks(
			app.TokenFactoryKeeper.Hooks(),
		))

	app.DexKeeper = *dexkeeper.NewKeeper(
		appCodec,
		keys[dextypes.StoreKey],
		keys[dextypes.MemStoreKey],
		app.BankKeeper.WithMintCoinsRestriction(dextypes.NewDexDenomMintCoinsRestriction()),
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	dexModule := dex.NewAppModule(appCodec, app.DexKeeper, app.BankKeeper)

	app.SwapKeeper = ibcswapkeeper.NewKeeper(
		appCodec,
		app.MsgServiceRouter(),
		app.IBCKeeper.ChannelKeeper,
		app.BankKeeper,
	)

	swapModule := ibcswap.NewAppModule(app.SwapKeeper)

	wasmDir := filepath.Join(homePath, "wasm")
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm cfg: %s", err))
	}

	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	// NOTE: we need staking feature here even if there is no staking module anymore because cosmwasm-std in the CosmWasm SDK requires this feature
	// NOTE: cosmwasm_1_2 feature enables GovMsg::VoteWeighted, which doesn't work with Neutron, because it uses its own custom governance,
	//       however, cosmwasm_1_2 also enables WasmMsg::Instantiate2, which works as one could expect
	supportedFeatures := "iterator,stargate,staking,neutron,cosmwasm_1_1,cosmwasm_1_2,cosmwasm_1_3,cosmwasm_1_4"

	// register the proposal types
	adminRouterLegacy := govv1beta1.NewRouter()
	adminRouterLegacy.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(&app.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper))

	app.AdminmoduleKeeper = *adminmodulekeeper.NewKeeper(
		appCodec,
		keys[adminmoduletypes.StoreKey],
		keys[adminmoduletypes.MemStoreKey],
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
		&app.BankKeeper,
		app.ContractManagerKeeper,
		interchainqueriesmodulekeeper.Verifier{},
		interchainqueriesmodulekeeper.TransactionVerifier{},
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)
	app.InterchainTxsKeeper = *interchaintxskeeper.NewKeeper(
		appCodec,
		keys[interchaintxstypes.StoreKey],
		memKeys[interchaintxstypes.MemStoreKey],
		app.IBCKeeper.ChannelKeeper,
		app.ICAControllerKeeper,
		contractmanager.NewSudoLimitWrapper(app.ContractManagerKeeper, &app.WasmKeeper),
		app.FeeKeeper,
		app.BankKeeper,
		func(ctx sdk.Context) string { return app.FeeBurnerKeeper.GetParams(ctx).TreasuryAddress },
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	app.CronKeeper = *cronkeeper.NewKeeper(
		appCodec,
		keys[crontypes.StoreKey],
		keys[crontypes.MemStoreKey],
		app.AccountKeeper,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)
	wasmOpts = append(wasmbinding.RegisterCustomPlugins(
		&app.InterchainTxsKeeper,
		&app.InterchainQueriesKeeper,
		app.TransferKeeper,
		&app.AdminmoduleKeeper,
		app.FeeBurnerKeeper,
		app.FeeKeeper, &app.BankKeeper,
		app.TokenFactoryKeeper, &app.CronKeeper,
		&app.ContractManagerKeeper,
		&app.DexKeeper,
	), wasmOpts...)

	queryPlugins := wasmkeeper.WithQueryPlugins(
		&wasmkeeper.QueryPlugins{Stargate: wasmkeeper.AcceptListStargateQuerier(wasmbinding.AcceptedStargateQueries(), app.GRPCQueryRouter(), appCodec)})
	wasmOpts = append(wasmOpts, queryPlugins)

	app.WasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		keys[wasmtypes.StoreKey],
		app.AccountKeeper,
		&app.BankKeeper,
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
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
		wasmOpts...,
	)
	wasmHooks.ContractKeeper = &app.WasmKeeper

	app.CronKeeper.WasmMsgServer = wasmkeeper.NewMsgServerImpl(&app.WasmKeeper)
	cronModule := cron.NewAppModule(appCodec, app.CronKeeper)

	transferIBCModule := transferSudo.NewIBCModule(
		app.TransferKeeper,
		contractmanager.NewSudoLimitWrapper(app.ContractManagerKeeper, &app.WasmKeeper),
	)
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

	app.PFMModule = packetforward.NewAppModule(app.PFMKeeper, app.GetSubspace(pfmtypes.ModuleName))

	var ibcStack ibcporttypes.IBCModule = packetforward.NewIBCMiddleware(
		app.HooksTransferIBCModule,
		app.PFMKeeper,
		0,
		pfmkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
		pfmkeeper.DefaultRefundTransferPacketTimeoutTimestamp,
	)

	ibcStack = ibcswap.NewIBCMiddleware(ibcStack, app.SwapKeeper)
	ibcStack = gmpmiddleware.NewIBCMiddleware(ibcStack)

	ibcRouter.AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(icahosttypes.SubModuleName, icaHostIBCModule).
		AddRoute(ibctransfertypes.ModuleName, ibcStack).
		AddRoute(interchaintxstypes.ModuleName, icaControllerStack).
		AddRoute(wasmtypes.ModuleName, wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper, app.IBCKeeper.ChannelKeeper)).
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
		app.PFMModule,
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
		swapModule,
		dexModule,
		// crisis always be last to make sure that it checks for all invariants and not only part of them
		crisis.NewAppModule(&app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator feerefunder pool to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
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
		wasmtypes.ModuleName,
		feetypes.ModuleName,
		feeburnertypes.ModuleName,
		adminmoduletypes.ModuleName,
		ibchookstypes.ModuleName,
		pfmtypes.ModuleName,
		crontypes.ModuleName,
		globalfee.ModuleName,
		ibcswaptypes.ModuleName,
		dextypes.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
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
		wasmtypes.ModuleName,
		feetypes.ModuleName,
		feeburnertypes.ModuleName,
		adminmoduletypes.ModuleName,
		ibchookstypes.ModuleName,
		pfmtypes.ModuleName,
		crontypes.ModuleName,
		globalfee.ModuleName,
		ibcswaptypes.ModuleName,
		// NOTE: Because of the gas sensitivity of PurgeExpiredLimit order operations
		// dexmodule must be the last endBlock module to run
		dextypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(
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
		wasmtypes.ModuleName,
		feetypes.ModuleName,
		feeburnertypes.ModuleName,
		adminmoduletypes.ModuleName,
		ibchookstypes.ModuleName, // after auth keeper
		pfmtypes.ModuleName,
		crontypes.ModuleName,
		globalfee.ModuleName,
		ibcswaptypes.ModuleName,
		dextypes.ModuleName,
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
		app.PFMModule,
		interchainQueriesModule,
		interchainTxsModule,
		feeBurnerModule,
		cronModule,
		dexModule,
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
			TXCounterStoreKey: keys[wasmtypes.StoreKey],
			ConsumerKeeper:    app.ConsumerKeeper,
			GlobalFeeSubspace: app.GetSubspace(globalfee.ModuleName),
		},
		app.Logger(),
	)
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)

	app.SetEndBlocker(app.EndBlocker)

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
		app.LoadLatest()
	}

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper
	app.ScopedWasmKeeper = scopedWasmKeeper
	app.ScopedInterTxKeeper = scopedInterTxKeeper
	app.ScopedCCVConsumerKeeper = scopedCCVConsumerKeeper

	return app
}

func (app *App) LoadLatest() {
	if err := app.LoadLatestVersion(); err != nil {
		tmos.Exit(err.Error())
	}

	ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})

	// Initialize pinned codes in wasmvm as they are not persisted there
	if err := app.WasmKeeper.InitializePinnedCodes(ctx); err != nil {
		tmos.Exit(fmt.Sprintf("failed initialize pinned codes %s", err))
	}
}

func (app *App) setupUpgradeStoreLoaders() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrd info from disk %s", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	for _, upgrd := range Upgrades {
		upgrd := upgrd
		if upgradeInfo.Name == upgrd.UpgradeName {
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &upgrd.StoreUpgrades))
		}
	}
}

func (app *App) setupUpgradeHandlers() {
	for _, upgrd := range Upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgrd.UpgradeName,
			upgrd.CreateUpgradeHandler(
				app.mm,
				app.configurator,
				&upgrades.UpgradeKeepers{
					AccountKeeper:       app.AccountKeeper,
					FeeBurnerKeeper:     app.FeeBurnerKeeper,
					CronKeeper:          app.CronKeeper,
					IcqKeeper:           app.InterchainQueriesKeeper,
					TokenFactoryKeeper:  app.TokenFactoryKeeper,
					SlashingKeeper:      app.SlashingKeeper,
					ParamsKeeper:        app.ParamsKeeper,
					CapabilityKeeper:    app.CapabilityKeeper,
					ContractManager:     app.ContractManagerKeeper,
					AdminModule:         app.AdminmoduleKeeper,
					ConsensusKeeper:     &app.ConsensusParamsKeeper,
					ConsumerKeeper:      &app.ConsumerKeeper,
					GlobalFeeSubspace:   app.GetSubspace(globalfee.ModuleName),
					CcvConsumerSubspace: app.GetSubspace(ccvconsumertypes.ModuleName),
				},
				app,
				app.AppCodec(),
			),
		)
	}
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

func (app *App) EnsureBlockGasMeter(ctx sdk.Context) {
	// TrancheKey generation and LimitOrderExpirationPurge both rely on a BlockGas meter.
	// check that it works at startup
	cp := app.GetConsensusParams(ctx)
	if cp == nil || cp.Block == nil || cp.Block.MaxGas <= 0 {
		panic("BlockGas meter must be initialized. Genesis must provide value for Block.MaxGas")
	}
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.EnsureBlockGasMeter(ctx)
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// InitChainer application update at chain initialization
// ONLY FOR TESTING PURPOSES
func (app *App) TestInitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	// manually set consensus params here, cause there is no way to set it using ibctesting stuff for now
	app.ConsensusParamsKeeper.Set(ctx, sims.DefaultConsensusParams)
	app.EnsureBlockGasMeter(ctx)

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

	paramsKeeper.Subspace(authtypes.ModuleName).WithKeyTable(authtypes.ParamKeyTable())         //nolint:staticcheck
	paramsKeeper.Subspace(banktypes.ModuleName).WithKeyTable(banktypes.ParamKeyTable())         //nolint:staticcheck
	paramsKeeper.Subspace(slashingtypes.ModuleName).WithKeyTable(slashingtypes.ParamKeyTable()) //nolint:staticcheck
	paramsKeeper.Subspace(crisistypes.ModuleName).WithKeyTable(crisistypes.ParamKeyTable())     //nolint:staticcheck
	paramsKeeper.Subspace(ibctransfertypes.ModuleName).WithKeyTable(ibctransfertypes.ParamKeyTable())
	paramsKeeper.Subspace(ibchost.ModuleName)
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName).WithKeyTable(icacontrollertypes.ParamKeyTable())
	paramsKeeper.Subspace(icahosttypes.SubModuleName).WithKeyTable(icahosttypes.ParamKeyTable())

	paramsKeeper.Subspace(pfmtypes.ModuleName).WithKeyTable(pfmtypes.ParamKeyTable())

	paramsKeeper.Subspace(globalfee.ModuleName).WithKeyTable(globalfeetypes.ParamKeyTable())

	paramsKeeper.Subspace(ccvconsumertypes.ModuleName).WithKeyTable(ccv.ParamKeyTable())

	// MOTE: legacy subspaces for migration sdk47 only
	paramsKeeper.Subspace(wasmtypes.ModuleName).WithKeyTable(wasmtypes.ParamKeyTable()) //nolint:staticcheck
	paramsKeeper.Subspace(crontypes.StoreKey).WithKeyTable(crontypes.ParamKeyTable())
	paramsKeeper.Subspace(feeburnertypes.StoreKey).WithKeyTable(feeburnertypes.ParamKeyTable())
	paramsKeeper.Subspace(feetypes.StoreKey).WithKeyTable(feetypes.ParamKeyTable())
	paramsKeeper.Subspace(tokenfactorytypes.StoreKey).WithKeyTable(tokenfactorytypes.ParamKeyTable())
	paramsKeeper.Subspace(interchainqueriesmoduletypes.StoreKey).WithKeyTable(interchainqueriesmoduletypes.ParamKeyTable())
	paramsKeeper.Subspace(interchaintxstypes.StoreKey).WithKeyTable(interchaintxstypes.ParamKeyTable())

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
func (app *App) GetStakingKeeper() ibctestingtypes.StakingKeeper {
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

// overrideWasmVariables overrides the wasm variables to:
//   - allow for larger wasm files
func overrideWasmVariables() {
	// Override Wasm size limitation from WASMD.
	wasmtypes.MaxWasmSize = 1_677_722 // ~1.6 mb (1024 * 1024 * 1.6)
	wasmtypes.MaxProposalWasmSize = wasmtypes.MaxWasmSize
}
