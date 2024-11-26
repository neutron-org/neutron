package app

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	v502 "github.com/neutron-org/neutron/v5/app/upgrades/v5.0.2"
	dynamicfeestypes "github.com/neutron-org/neutron/v5/x/dynamicfees/types"

	"github.com/skip-mev/feemarket/x/feemarket"
	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"

	"github.com/neutron-org/neutron/v5/x/dynamicfees"
	ibcratelimit "github.com/neutron-org/neutron/v5/x/ibc-rate-limit"

	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	ethcryptocodec "github.com/neutron-org/neutron/v5/x/crypto/codec"

	appconfig "github.com/neutron-org/neutron/v5/app/config"

	"github.com/skip-mev/slinky/abci/strategies/aggregator"
	"github.com/skip-mev/slinky/x/oracle"

	oraclepreblock "github.com/skip-mev/slinky/abci/preblock/oracle"
	slinkyproposals "github.com/skip-mev/slinky/abci/proposals"
	compression "github.com/skip-mev/slinky/abci/strategies/codec"
	"github.com/skip-mev/slinky/abci/strategies/currencypair"
	"github.com/skip-mev/slinky/abci/ve"
	oracleconfig "github.com/skip-mev/slinky/oracle/config"
	"github.com/skip-mev/slinky/pkg/math/voteweighted"
	oracleclient "github.com/skip-mev/slinky/service/clients/oracle"
	servicemetrics "github.com/skip-mev/slinky/service/metrics"

	v500 "github.com/neutron-org/neutron/v5/app/upgrades/v5.0.0"
	"github.com/neutron-org/neutron/v5/x/globalfee"
	globalfeetypes "github.com/neutron-org/neutron/v5/x/globalfee/types"

	"cosmossdk.io/log"
	db "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec/address"

	// globalfeetypes "github.com/cosmos/gaia/v11/x/globalfee/types"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward"
	ibctestingtypes "github.com/cosmos/ibc-go/v8/testing/types"
	"github.com/cosmos/interchain-security/v5/testutil/integration"
	ccv "github.com/cosmos/interchain-security/v5/x/ccv/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	tendermint "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"

	"github.com/neutron-org/neutron/v5/docs"

	"github.com/neutron-org/neutron/v5/app/upgrades"

	"github.com/neutron-org/neutron/v5/x/cron"

	"cosmossdk.io/x/evidence"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	tmos "github.com/cometbft/cometbft/libs/os"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
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
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	// "github.com/cosmos/gaia/v11/x/globalfee"
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"
	icacontroller "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahost "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:staticcheck
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"

	ibcratelimitkeeper "github.com/neutron-org/neutron/v5/x/ibc-rate-limit/keeper"
	ibcratelimittypes "github.com/neutron-org/neutron/v5/x/ibc-rate-limit/types"

	//nolint:staticcheck
	ibcporttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/spf13/cast"

	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	cronkeeper "github.com/neutron-org/neutron/v5/x/cron/keeper"
	crontypes "github.com/neutron-org/neutron/v5/x/cron/types"

	"github.com/neutron-org/neutron/v5/x/tokenfactory"
	tokenfactorykeeper "github.com/neutron-org/neutron/v5/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/neutron-org/neutron/v5/x/tokenfactory/types"

	"github.com/cosmos/admin-module/v2/x/adminmodule"
	adminmodulecli "github.com/cosmos/admin-module/v2/x/adminmodule/client/cli"
	adminmodulekeeper "github.com/cosmos/admin-module/v2/x/adminmodule/keeper"
	adminmoduletypes "github.com/cosmos/admin-module/v2/x/adminmodule/types"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	appparams "github.com/neutron-org/neutron/v5/app/params"
	"github.com/neutron-org/neutron/v5/wasmbinding"
	"github.com/neutron-org/neutron/v5/x/contractmanager"
	contractmanagermodulekeeper "github.com/neutron-org/neutron/v5/x/contractmanager/keeper"
	contractmanagermoduletypes "github.com/neutron-org/neutron/v5/x/contractmanager/types"
	dynamicfeeskeeper "github.com/neutron-org/neutron/v5/x/dynamicfees/keeper"
	"github.com/neutron-org/neutron/v5/x/feeburner"
	feeburnerkeeper "github.com/neutron-org/neutron/v5/x/feeburner/keeper"
	feeburnertypes "github.com/neutron-org/neutron/v5/x/feeburner/types"
	"github.com/neutron-org/neutron/v5/x/feerefunder"
	feekeeper "github.com/neutron-org/neutron/v5/x/feerefunder/keeper"
	ibchooks "github.com/neutron-org/neutron/v5/x/ibc-hooks"
	ibchookstypes "github.com/neutron-org/neutron/v5/x/ibc-hooks/types"
	"github.com/neutron-org/neutron/v5/x/interchainqueries"
	interchainqueriesmodulekeeper "github.com/neutron-org/neutron/v5/x/interchainqueries/keeper"
	interchainqueriesmoduletypes "github.com/neutron-org/neutron/v5/x/interchainqueries/types"
	"github.com/neutron-org/neutron/v5/x/interchaintxs"
	interchaintxskeeper "github.com/neutron-org/neutron/v5/x/interchaintxs/keeper"
	interchaintxstypes "github.com/neutron-org/neutron/v5/x/interchaintxs/types"
	transferSudo "github.com/neutron-org/neutron/v5/x/transfer"
	wrapkeeper "github.com/neutron-org/neutron/v5/x/transfer/keeper"

	feetypes "github.com/neutron-org/neutron/v5/x/feerefunder/types"

	ccvconsumer "github.com/cosmos/interchain-security/v5/x/ccv/consumer"
	ccvconsumerkeeper "github.com/cosmos/interchain-security/v5/x/ccv/consumer/keeper"
	ccvconsumertypes "github.com/cosmos/interchain-security/v5/x/ccv/consumer/types"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	pfmkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/keeper"
	pfmtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"

	"github.com/neutron-org/neutron/v5/x/dex"
	dexkeeper "github.com/neutron-org/neutron/v5/x/dex/keeper"
	dextypes "github.com/neutron-org/neutron/v5/x/dex/types"

	globalfeekeeper "github.com/neutron-org/neutron/v5/x/globalfee/keeper"
	gmpmiddleware "github.com/neutron-org/neutron/v5/x/gmp"

	// Block-sdk imports
	blocksdkabci "github.com/skip-mev/block-sdk/v2/abci"
	blocksdk "github.com/skip-mev/block-sdk/v2/block"
	"github.com/skip-mev/block-sdk/v2/x/auction"
	auctionkeeper "github.com/skip-mev/block-sdk/v2/x/auction/keeper"
	rewardsaddressprovider "github.com/skip-mev/block-sdk/v2/x/auction/rewards"
	auctiontypes "github.com/skip-mev/block-sdk/v2/x/auction/types"

	"github.com/skip-mev/block-sdk/v2/abci/checktx"
	"github.com/skip-mev/block-sdk/v2/block/base"

	"github.com/skip-mev/slinky/x/marketmap"
	marketmapkeeper "github.com/skip-mev/slinky/x/marketmap/keeper"
	marketmaptypes "github.com/skip-mev/slinky/x/marketmap/types"
	oraclekeeper "github.com/skip-mev/slinky/x/oracle/keeper"
	oracletypes "github.com/skip-mev/slinky/x/oracle/types"

	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
)

const (
	Name = "neutrond"
)

var (
	Upgrades = []upgrades.Upgrade{
		v500.Upgrade,
		v502.Upgrade,
	}

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
		ibcratelimit.AppModuleBasic{},
		auction.AppModuleBasic{},
		globalfee.AppModule{},
		feemarket.AppModuleBasic{},
		dex.AppModuleBasic{},
		oracle.AppModuleBasic{},
		marketmap.AppModuleBasic{},
		dynamicfees.AppModuleBasic{},
		consensus.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:                    nil,
		auctiontypes.ModuleName:                       nil,
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
		oracletypes.ModuleName:                        nil,
		marketmaptypes.ModuleName:                     nil,
		feemarkettypes.FeeCollectorName:               nil,
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

	appconfig.GetDefaultConfig()
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
	// AuctionKeeper handles the processing of bid-txs, the selection of winners per height, and the distribution of rewards.
	AuctionKeeper       auctionkeeper.Keeper
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
	FeeMarkerKeeper     *feemarketkeeper.Keeper
	DynamicFeesKeeper   *dynamicfeeskeeper.Keeper
	FeeKeeper           *feekeeper.Keeper
	FeeBurnerKeeper     *feeburnerkeeper.Keeper
	ConsumerKeeper      ccvconsumerkeeper.Keeper
	TokenFactoryKeeper  *tokenfactorykeeper.Keeper
	CronKeeper          cronkeeper.Keeper
	PFMKeeper           *pfmkeeper.Keeper
	DexKeeper           dexkeeper.Keeper
	GlobalFeeKeeper     globalfeekeeper.Keeper

	PFMModule packetforward.AppModule

	TransferStack           *ibchooks.IBCMiddleware
	Ics20WasmHooks          *ibchooks.WasmHooks
	RateLimitingICS4Wrapper *ibcratelimit.ICS4Wrapper
	HooksICS4Wrapper        ibchooks.ICS4Middleware

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

	WasmKeeper     wasmkeeper.Keeper
	ContractKeeper *wasmkeeper.PermissionedKeeper

	// slinky
	MarketMapKeeper       *marketmapkeeper.Keeper
	OracleKeeper          *oraclekeeper.Keeper
	oraclePreBlockHandler *oraclepreblock.PreBlockHandler

	// processes
	oracleClient oracleclient.OracleClient

	// mm is the module manager
	mm *module.Manager

	// sm is the simulation manager
	sm *module.SimulationManager

	// Custom checkTx handler -> this check-tx is used to simulate txs that are
	// wrapped in a bid-tx
	checkTxHandler checktx.CheckTx
}

// AutoCLIOpts returns options based upon the modules in the neutron v5 app.
func (app *App) AutoCLIOpts(initClientCtx client.Context) autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule)
	for _, m := range app.mm.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.mm.Modules),
		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
		ClientCtx:             initClientCtx,
	}
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

func (app *App) GetTestEvidenceKeeper() evidencekeeper.Keeper {
	return app.EvidenceKeeper
}

// New returns a reference to an initialized blockchain app
func New(
	logger log.Logger,
	db db.DB,
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

	//ethcryptocodec.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ethcryptocodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	bApp := baseapp.NewBaseApp(Name, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := storetypes.NewKVStoreKeys(
		authzkeeper.StoreKey, authtypes.StoreKey, banktypes.StoreKey, slashingtypes.StoreKey,
		paramstypes.StoreKey, ibchost.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey, icacontrollertypes.StoreKey,
		icahosttypes.StoreKey, capabilitytypes.StoreKey,
		interchainqueriesmoduletypes.StoreKey, contractmanagermoduletypes.StoreKey, interchaintxstypes.StoreKey, wasmtypes.StoreKey, feetypes.StoreKey,
		feeburnertypes.StoreKey, adminmoduletypes.StoreKey, ccvconsumertypes.StoreKey, tokenfactorytypes.StoreKey, pfmtypes.StoreKey,
		crontypes.StoreKey, ibcratelimittypes.ModuleName, ibchookstypes.StoreKey, consensusparamtypes.StoreKey, crisistypes.StoreKey, dextypes.StoreKey, auctiontypes.StoreKey,
		oracletypes.StoreKey, marketmaptypes.StoreKey, feemarkettypes.StoreKey, dynamicfeestypes.StoreKey, globalfeetypes.StoreKey,
	)
	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey, dextypes.TStoreKey)
	memKeys := storetypes.NewMemoryStoreKeys(capabilitytypes.MemStoreKey, feetypes.MemStoreKey)

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
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]), authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(), runtime.EventService{})
	bApp.SetParamStore(&app.ConsensusParamsKeeper.ParamsStore)

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])

	// grant capabilities for the ibc and ibc-transfer modules
	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	scopedICAControllerKeeper := app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	app.ScopedTransferKeeper = scopedTransferKeeper
	scopedWasmKeeper := app.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)
	scopedInterTxKeeper := app.CapabilityKeeper.ScopeToModule(interchaintxstypes.ModuleName)
	scopedCCVConsumerKeeper := app.CapabilityKeeper.ScopeToModule(ccvconsumertypes.ModuleName)

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		address.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[authz.ModuleName]), appCodec, app.MsgServiceRouter(), app.AccountKeeper,
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		app.BlockedAddrs(),
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
		logger,
	)

	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		runtime.NewKVStoreService(keys[slashingtypes.StoreKey]),
		&app.ConsumerKeeper,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)
	app.CrisisKeeper = *crisiskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[crisistypes.StoreKey]),
		invCheckPeriod,
		&app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
		address.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[feegrant.StoreKey]), app.AccountKeeper)
	app.UpgradeKeeper = *upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		appCodec,
		homePath,
		app.BaseApp,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	app.DynamicFeesKeeper = dynamicfeeskeeper.NewKeeper(appCodec, keys[dynamicfeestypes.StoreKey], authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String())

	app.FeeMarkerKeeper = feemarketkeeper.NewKeeper(
		appCodec,
		keys[feemarkettypes.StoreKey],
		app.AccountKeeper,
		app.DynamicFeesKeeper,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	// ... other modules keepers
	// pre-initialize ConsumerKeeper to satisfy ibckeeper.NewKeeper
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
		appCodec, keys[ibchost.StoreKey], app.GetSubspace(ibchost.ModuleName), &app.ConsumerKeeper, app.UpgradeKeeper, scopedIBCKeeper, authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	// Feekeeper needs to be initialized before middlewares injection
	app.FeeKeeper = feekeeper.NewKeeper(
		appCodec,
		keys[feetypes.StoreKey],
		memKeys[feetypes.MemStoreKey],
		app.IBCKeeper.ChannelKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)
	feeModule := feerefunder.NewAppModule(appCodec, *app.FeeKeeper, app.AccountKeeper, app.BankKeeper)

	app.ContractManagerKeeper = *contractmanagermodulekeeper.NewKeeper(
		appCodec,
		keys[contractmanagermoduletypes.StoreKey],
		keys[contractmanagermoduletypes.MemStoreKey],
		&app.WasmKeeper,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	app.WireICS20PreWasmKeeper(appCodec)
	app.PFMModule = packetforward.NewAppModule(app.PFMKeeper, app.GetSubspace(pfmtypes.ModuleName))

	app.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		appCodec, keys[icacontrollertypes.StoreKey], app.GetSubspace(icacontrollertypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper, app.IBCKeeper.PortKeeper,
		scopedICAControllerKeeper, app.MsgServiceRouter(),
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec, keys[icahosttypes.StoreKey], app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper, app.IBCKeeper.PortKeeper,
		app.AccountKeeper, scopedICAHostKeeper, app.MsgServiceRouter(),
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)
	app.ICAHostKeeper.WithQueryRouter(app.GRPCQueryRouter())

	app.FeeBurnerKeeper = feeburnerkeeper.NewKeeper(
		appCodec,
		keys[feeburnertypes.StoreKey],
		keys[feeburnertypes.MemStoreKey],
		app.AccountKeeper,
		&app.BankKeeper,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)
	feeBurnerModule := feeburner.NewAppModule(appCodec, *app.FeeBurnerKeeper)

	app.GlobalFeeKeeper = globalfeekeeper.NewKeeper(appCodec, keys[globalfeetypes.StoreKey], authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String())

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(keys[evidencetypes.StoreKey]), &app.ConsumerKeeper, app.SlashingKeeper,
		address.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()), runtime.ProvideCometInfoService(),
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	app.ConsumerKeeper = ccvconsumerkeeper.NewKeeper(
		appCodec,
		keys[ccvconsumertypes.StoreKey],
		app.GetSubspace(ccvconsumertypes.ModuleName),
		scopedCCVConsumerKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.IBCKeeper.ConnectionKeeper,
		app.IBCKeeper.ClientKeeper,
		app.SlashingKeeper,
		&app.BankKeeper,
		app.AccountKeeper,
		app.TransferKeeper.Keeper, // we cant use our transfer wrapper type here because of interface incompatibility, it looks safe to use underlying transfer keeper.
		// Since the keeper is only used to send reward to provider chain
		app.IBCKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
		address.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		address.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)
	app.ConsumerKeeper = *app.ConsumerKeeper.SetHooks(app.SlashingKeeper.Hooks())
	consumerModule := ccvconsumer.NewAppModule(app.ConsumerKeeper, app.GetSubspace(ccvconsumertypes.ModuleName))

	tokenFactoryKeeper := tokenfactorykeeper.NewKeeper(
		appCodec,
		app.keys[tokenfactorytypes.StoreKey],
		maccPerms,
		app.AccountKeeper,
		&app.BankKeeper,
		&app.WasmKeeper,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)
	app.TokenFactoryKeeper = &tokenFactoryKeeper

	app.BankKeeper.BaseSendKeeper = app.BankKeeper.BaseSendKeeper.SetHooks(
		banktypes.NewMultiBankHooks(
			app.TokenFactoryKeeper.Hooks(),
		))

	app.DexKeeper = *dexkeeper.NewKeeper(
		appCodec,
		keys[dextypes.StoreKey],
		keys[dextypes.MemStoreKey],
		tkeys[dextypes.TStoreKey],
		app.BankKeeper.WithMintCoinsRestriction(dextypes.NewDexDenomMintCoinsRestriction()),
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	app.AuctionKeeper = auctionkeeper.NewKeeperWithRewardsAddressProvider(
		appCodec,
		keys[auctiontypes.StoreKey],
		app.AccountKeeper,
		&app.BankKeeper,
		// 25% of rewards should be sent to the redistribute address
		rewardsaddressprovider.NewFixedAddressRewardsAddressProvider(app.AccountKeeper.GetModuleAddress(ccvconsumertypes.ConsumerRedistributeName)),
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	dexModule := dex.NewAppModule(appCodec, app.DexKeeper, app.BankKeeper)

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
	supportedFeatures := []string{"iterator", "stargate", "staking", "neutron", "cosmwasm_1_1", "cosmwasm_1_2", "cosmwasm_1_3", "cosmwasm_1_4", "cosmwasm_2_0", "cosmwasm_2_1"}

	// register the proposal types
	adminRouterLegacy := govv1beta1.NewRouter()
	adminRouterLegacy.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)) //nolint:staticcheck

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
		icacontrollerkeeper.NewMsgServerImpl(&app.ICAControllerKeeper),
		contractmanager.NewSudoLimitWrapper(app.ContractManagerKeeper, &app.WasmKeeper),
		app.FeeKeeper,
		app.BankKeeper,
		func(ctx sdk.Context) string { return app.FeeBurnerKeeper.GetParams(ctx).TreasuryAddress },
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	app.MarketMapKeeper = marketmapkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[marketmaptypes.StoreKey]),
		appCodec,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName),
	)
	marketmapModule := marketmap.NewAppModule(appCodec, app.MarketMapKeeper)

	oracleKeeper := oraclekeeper.NewKeeper(runtime.NewKVStoreService(keys[oracletypes.StoreKey]),
		appCodec,
		app.MarketMapKeeper,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName))
	app.OracleKeeper = &oracleKeeper
	oracleModule := oracle.NewAppModule(appCodec, *app.OracleKeeper)

	app.MarketMapKeeper.SetHooks(app.OracleKeeper.Hooks())

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
		app.FeeKeeper,
		&app.BankKeeper,
		app.TokenFactoryKeeper,
		&app.CronKeeper,
		&app.ContractManagerKeeper,
		&app.DexKeeper,
		app.OracleKeeper,
		app.MarketMapKeeper,
	), wasmOpts...)

	queryPlugins := wasmkeeper.WithQueryPlugins(
		&wasmkeeper.QueryPlugins{
			Stargate: wasmkeeper.AcceptListStargateQuerier(wasmbinding.AcceptedStargateQueries(), app.GRPCQueryRouter(), appCodec),
			Grpc:     wasmkeeper.AcceptListGrpcQuerier(wasmbinding.AcceptedStargateQueries(), app.GRPCQueryRouter(), appCodec),
		})
	wasmOpts = append(wasmOpts, queryPlugins)

	app.WasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[wasmtypes.StoreKey]),
		app.AccountKeeper,
		&app.BankKeeper,
		nil,
		nil,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
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

	app.CronKeeper.WasmMsgServer = wasmkeeper.NewMsgServerImpl(&app.WasmKeeper)
	cronModule := cron.NewAppModule(appCodec, app.CronKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := ibcporttypes.NewRouter()

	icaModule := ica.NewAppModule(&app.ICAControllerKeeper, &app.ICAHostKeeper)

	var icaControllerStack ibcporttypes.IBCModule

	icaControllerStack = interchaintxs.NewIBCModule(app.InterchainTxsKeeper)
	icaControllerStack = icacontroller.NewIBCMiddleware(icaControllerStack, app.ICAControllerKeeper)

	icaHostIBCModule := icahost.NewIBCModule(app.ICAHostKeeper)

	interchainQueriesModule := interchainqueries.NewAppModule(
		appCodec,
		keys[interchainqueriesmoduletypes.StoreKey],
		app.InterchainQueriesKeeper,
		app.AccountKeeper,
		app.BankKeeper,
	)
	interchainTxsModule := interchaintxs.NewAppModule(appCodec, app.InterchainTxsKeeper, app.AccountKeeper, app.BankKeeper)
	contractManagerModule := contractmanager.NewAppModule(appCodec, app.ContractManagerKeeper)
	ibcRateLimitmodule := ibcratelimit.NewAppModule(appCodec, app.RateLimitingICS4Wrapper.IbcratelimitKeeper, app.RateLimitingICS4Wrapper)
	ibcHooksModule := ibchooks.NewAppModule(app.AccountKeeper)

	transferModule := transferSudo.NewAppModule(app.TransferKeeper)
	app.ContractKeeper = wasmkeeper.NewDefaultPermissionKeeper(app.WasmKeeper)

	app.RateLimitingICS4Wrapper.ContractKeeper = app.ContractKeeper
	app.Ics20WasmHooks.ContractKeeper = &app.WasmKeeper

	ibcRouter.AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
		AddRoute(icahosttypes.SubModuleName, icaHostIBCModule).
		AddRoute(ibctransfertypes.ModuleName, app.TransferStack).
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
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.ConsumerKeeper, app.GetSubspace(slashingtypes.ModuleName), app.interfaceRegistry),
		upgrade.NewAppModule(&app.UpgradeKeeper, address.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())),
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
		ibcRateLimitmodule,
		ibcHooksModule,
		tokenfactory.NewAppModule(appCodec, *app.TokenFactoryKeeper, app.AccountKeeper, app.BankKeeper),
		cronModule,
		globalfee.NewAppModule(app.GlobalFeeKeeper, app.GetSubspace(globalfee.ModuleName), app.AppCodec(), app.keys[globalfee.ModuleName]),
		feemarket.NewAppModule(appCodec, *app.FeeMarkerKeeper),
		dynamicfees.NewAppModule(appCodec, *app.DynamicFeesKeeper),
		dexModule,
		marketmapModule,
		oracleModule,
		auction.NewAppModule(appCodec, app.AuctionKeeper),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		// always be last to make sure that it checks for all invariants and not only part of them
		crisis.NewAppModule(&app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
	)

	app.mm.SetOrderPreBlockers(
		upgradetypes.ModuleName,
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator feerefunder pool to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		auctiontypes.ModuleName,
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
		ibcratelimittypes.ModuleName,
		ibchookstypes.ModuleName,
		pfmtypes.ModuleName,
		crontypes.ModuleName,
		marketmaptypes.ModuleName,
		oracletypes.ModuleName,
		globalfee.ModuleName,
		feemarkettypes.ModuleName,
		dextypes.ModuleName,
		consensusparamtypes.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
		auctiontypes.ModuleName,
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
		ibcratelimittypes.ModuleName,
		ibchookstypes.ModuleName,
		pfmtypes.ModuleName,
		crontypes.ModuleName,
		marketmaptypes.ModuleName,
		oracletypes.ModuleName,
		globalfee.ModuleName,
		feemarkettypes.ModuleName,
		dextypes.ModuleName,
		consensusparamtypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(
		auctiontypes.ModuleName,
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
		ibcratelimittypes.ModuleName,
		ibchookstypes.ModuleName, // after auth keeper
		pfmtypes.ModuleName,
		crontypes.ModuleName,
		globalfee.ModuleName,
		feemarkettypes.ModuleName,
		oracletypes.ModuleName,
		marketmaptypes.ModuleName,
		dextypes.ModuleName,
		dynamicfeestypes.ModuleName,
		consensusparamtypes.ModuleName,
	)

	app.mm.RegisterInvariants(&app.CrisisKeeper)
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	err = app.mm.RegisterServices(app.configurator)
	if err != nil {
		panic(fmt.Sprintf("failed to register services: %s", err))
	}

	app.setupUpgradeHandlers()

	// create the simulation manager and define the order of the modules for deterministic simulations
	app.sm = module.NewSimulationManager(
		auth.NewAppModule(appCodec, app.AccountKeeper, nil, app.GetSubspace(authtypes.ModuleName)),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, nil, app.GetSubspace(slashingtypes.ModuleName), app.interfaceRegistry),
		wasm.NewAppModule(appCodec, &app.WasmKeeper, app.AccountKeeper, app.BankKeeper, app.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
		evidence.NewAppModule(app.EvidenceKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		params.NewAppModule(app.ParamsKeeper),
		transferModule,
		ibcRateLimitmodule,
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
	app.SetEndBlocker(app.EndBlocker)

	// create the lanes
	baseLane := app.CreateLanes()
	mempool, err := blocksdk.NewLanedMempool(app.Logger(), []blocksdk.Lane{baseLane})
	if err != nil {
		panic(err)
	}

	// set the mempool first
	app.SetMempool(mempool)

	// then create the ante-handler
	anteHandler, err := NewAnteHandler(
		HandlerOptions{
			HandlerOptions: ante.HandlerOptions{
				FeegrantKeeper:  app.FeeGrantKeeper,
				SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
				SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			},
			BankKeeper:            app.BankKeeper,
			AccountKeeper:         app.AccountKeeper,
			IBCKeeper:             app.IBCKeeper,
			GlobalFeeKeeper:       app.GlobalFeeKeeper,
			WasmConfig:            &wasmConfig,
			TXCounterStoreService: runtime.NewKVStoreService(keys[wasmtypes.StoreKey]),
			ConsumerKeeper:        app.ConsumerKeeper,
			FeeMarketKeeper:       app.FeeMarkerKeeper,
		},
		app.Logger(),
	)
	if err != nil {
		panic(err)
	}
	app.SetAnteHandler(anteHandler)

	postHandlerOptions := PostHandlerOptions{
		AccountKeeper:   app.AccountKeeper,
		BankKeeper:      app.BankKeeper,
		FeeMarketKeeper: app.FeeMarkerKeeper,
	}
	postHandler, err := NewPostHandler(postHandlerOptions)
	if err != nil {
		panic(err)
	}
	app.SetPostHandler(postHandler)

	// set ante-handlers
	opts := []base.LaneOption{
		base.WithAnteHandler(anteHandler),
	}
	baseLane.WithOptions(opts...)

	// set the block-sdk prepare / process-proposal handlers
	blockSdkProposalHandler := blocksdkabci.NewDefaultProposalHandler(
		app.Logger(),
		app.GetTxConfig().TxDecoder(),
		app.GetTxConfig().TxEncoder(),
		mempool,
	)

	// Read general config from app-opts, and construct oracle service.
	cfg, err := oracleconfig.ReadConfigFromAppOpts(appOpts)
	if err != nil {
		panic(err)
	}

	// If app level instrumentation is enabled, then wrap the oracle service with a metrics client
	// to get metrics on the oracle service (for ABCI++). This will allow the instrumentation to track
	// latency in VerifyVoteExtension requests and more.
	oracleMetrics, err := servicemetrics.NewMetricsFromConfig(cfg, app.ChainID())
	if err != nil {
		panic(err)
	}

	// Create the oracle service.
	app.oracleClient, err = oracleclient.NewClientFromConfig(
		cfg,
		app.Logger().With("client", "oracle"),
		oracleMetrics,
	)
	if err != nil {
		panic(err)
	}

	// Connect to the oracle service
	if err := app.oracleClient.Start(context.Background()); err != nil {
		app.Logger().Error("failed to start oracle client", "err", err)
		panic(err)
	}

	// Create special kind of store to implement ValidatorStore interfaces for ConsumerKeeper (as we don't have StakingKeeper)
	ccvconsumerCompatKeeper := voteweighted.NewCCVConsumerCompatKeeper(app.ConsumerKeeper)

	// Create the proposal handler that will be used to fill proposals with
	// transactions and oracle data.
	oracleProposalHandler := slinkyproposals.NewProposalHandler(
		app.Logger(),
		blockSdkProposalHandler.PrepareProposalHandler(),
		baseapp.NoOpProcessProposal(),
		ve.NewDefaultValidateVoteExtensionsFn(ccvconsumerCompatKeeper),
		compression.NewCompressionVoteExtensionCodec(
			compression.NewDefaultVoteExtensionCodec(),
			compression.NewZLibCompressor(),
		),
		compression.NewCompressionExtendedCommitCodec(
			compression.NewDefaultExtendedCommitCodec(),
			compression.NewZStdCompressor(),
		),
		currencypair.NewDeltaCurrencyPairStrategy(app.OracleKeeper),
		oracleMetrics,
	)
	app.SetPrepareProposal(oracleProposalHandler.PrepareProposalHandler())
	app.SetProcessProposal(oracleProposalHandler.ProcessProposalHandler())

	// wrap checkTxHandler with mempool parity handler
	parityCheckTx := checktx.NewMempoolParityCheckTx(
		app.Logger(),
		mempool,
		app.GetTxConfig().TxDecoder(),
		app.BaseApp.CheckTx,
		app.BaseApp,
	)

	app.SetCheckTx(parityCheckTx.CheckTx())

	// Create the aggregation function that will be used to aggregate oracle data
	// from each validator.
	aggregatorFn := voteweighted.MedianFromContext(
		app.Logger(),
		ccvconsumerCompatKeeper,
		voteweighted.DefaultPowerThreshold,
	)

	// Create the pre-finalize block hook that will be used to apply oracle data
	// to the state before any transactions are executed (in finalize block).
	app.oraclePreBlockHandler = oraclepreblock.NewOraclePreBlockHandler(
		app.Logger(),
		aggregatorFn,
		app.OracleKeeper,
		oracleMetrics,
		currencypair.NewDeltaCurrencyPairStrategy(app.OracleKeeper),
		compression.NewCompressionVoteExtensionCodec(
			compression.NewDefaultVoteExtensionCodec(),
			compression.NewZLibCompressor(),
		),
		compression.NewCompressionExtendedCommitCodec(
			compression.NewDefaultExtendedCommitCodec(),
			compression.NewZStdCompressor(),
		),
	)

	app.SetPreBlocker(app.oraclePreBlockHandler.WrappedPreBlocker(app.mm))

	// Create the vote extensions handler that will be used to extend and verify
	// vote extensions (i.e. oracle data).
	veCodec := compression.NewCompressionVoteExtensionCodec(
		compression.NewDefaultVoteExtensionCodec(),
		compression.NewZLibCompressor(),
	)
	extCommitCodec := compression.NewCompressionExtendedCommitCodec(
		compression.NewDefaultExtendedCommitCodec(),
		compression.NewZStdCompressor(),
	)
	voteExtensionsHandler := ve.NewVoteExtensionHandler(
		app.Logger(),
		app.oracleClient,
		time.Second,
		currencypair.NewDeltaCurrencyPairStrategy(app.OracleKeeper),
		compression.NewCompressionVoteExtensionCodec(
			compression.NewDefaultVoteExtensionCodec(),
			compression.NewZLibCompressor(),
		),
		aggregator.NewOraclePriceApplier(
			aggregator.NewDefaultVoteAggregator(
				app.Logger(),
				aggregatorFn,
				currencypair.NewDeltaCurrencyPairStrategy(app.OracleKeeper),
			),
			app.OracleKeeper,
			veCodec,
			extCommitCodec,
			app.Logger(),
		),
		oracleMetrics,
	)
	app.SetExtendVoteHandler(voteExtensionsHandler.ExtendVoteHandler())
	app.SetVerifyVoteExtensionHandler(voteExtensionsHandler.VerifyVoteExtensionHandler())

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
					AuctionKeeper:       app.AuctionKeeper,
					ContractManager:     app.ContractManagerKeeper,
					AdminModule:         app.AdminmoduleKeeper,
					ConsensusKeeper:     &app.ConsensusParamsKeeper,
					ConsumerKeeper:      &app.ConsumerKeeper,
					MarketmapKeeper:     app.MarketMapKeeper,
					FeeMarketKeeper:     app.FeeMarkerKeeper,
					DynamicfeesKeeper:   app.DynamicFeesKeeper,
					DexKeeper:           &app.DexKeeper,
					IbcRateLimitKeeper:  app.RateLimitingICS4Wrapper.IbcratelimitKeeper,
					GlobalFeeSubspace:   app.GetSubspace(globalfee.ModuleName),
					CcvConsumerSubspace: app.GetSubspace(ccvconsumertypes.ModuleName),
				},
				app,
				app.AppCodec(),
			),
		)
	}
}

func (app *App) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)
	for _, m := range app.mm.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.mm.Modules),
		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}
}

// CheckTx will check the transaction with the provided checkTxHandler. We override the default
// handler so that we can verify bid transactions before they are inserted into the mempool.
// With the Block-SDK CheckTx, we can verify the bid transaction and all of the bundled transactions
// before inserting the bid transaction into the mempool.
func (app *App) CheckTx(req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	return app.checkTxHandler(req)
}

// SetCheckTx sets the checkTxHandler for the app.
func (app *App) SetCheckTx(handler checktx.CheckTx) {
	app.checkTxHandler = handler
}

// Name returns the name of the App
func (app *App) Name() string { return app.BaseApp.Name() }

// GetBaseApp returns the base app of the application
func (app *App) GetBaseApp() *baseapp.BaseApp { return app.BaseApp }

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.mm.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *App) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.mm.EndBlock(ctx)
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		return nil, err
	}
	if err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap()); err != nil {
		return nil, err
	}
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// InitChainer application update at chain initialization
// ONLY FOR TESTING PURPOSES
func (app *App) TestInitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	// manually set consensus params here, cause there is no way to set it using ibctesting stuff for now
	// TODO: app.ConsensusParamsKeeper.Set(ctx, sims.DefaultConsensusParams)

	err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	if err != nil {
		return nil, fmt.Errorf("failed to set module version map: %w", err)
	}
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
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

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
	cmtservice.RegisterTendermintService(clientCtx, app.BaseApp.GRPCQueryRouter(), app.interfaceRegistry, app.Query)
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName).WithKeyTable(authtypes.ParamKeyTable())         //nolint:staticcheck
	paramsKeeper.Subspace(banktypes.ModuleName).WithKeyTable(banktypes.ParamKeyTable())         //nolint:staticcheck
	paramsKeeper.Subspace(slashingtypes.ModuleName).WithKeyTable(slashingtypes.ParamKeyTable()) //nolint:staticcheck
	paramsKeeper.Subspace(crisistypes.ModuleName).WithKeyTable(crisistypes.ParamKeyTable())     //nolint:staticcheck
	paramsKeeper.Subspace(ibctransfertypes.ModuleName).WithKeyTable(ibctransfertypes.ParamKeyTable())

	keyTable := ibcclienttypes.ParamKeyTable()
	keyTable.RegisterParamSet(&ibcconnectiontypes.Params{})
	paramsKeeper.Subspace(ibchost.ModuleName).WithKeyTable(keyTable)

	paramsKeeper.Subspace(icacontrollertypes.SubModuleName).WithKeyTable(icacontrollertypes.ParamKeyTable())
	paramsKeeper.Subspace(icahosttypes.SubModuleName).WithKeyTable(icahosttypes.ParamKeyTable())

	paramsKeeper.Subspace(pfmtypes.ModuleName).WithKeyTable(pfmtypes.ParamKeyTable())

	paramsKeeper.Subspace(globalfee.ModuleName).WithKeyTable(globalfeetypes.ParamKeyTable())

	paramsKeeper.Subspace(ccvconsumertypes.ModuleName).WithKeyTable(ccv.ParamKeyTable())

	// MOTE: legacy subspaces for migration sdk47 only //nolint:staticcheck
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

func (app *App) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// overrideWasmVariables overrides the wasm variables to:
//   - allow for larger wasm files
func overrideWasmVariables() {
	// Override Wasm size limitation from WASMD.
	wasmtypes.MaxWasmSize = 1_677_722 // ~1.6 mb (1024 * 1024 * 1.6)
	wasmtypes.MaxProposalWasmSize = wasmtypes.MaxWasmSize
}

// WireICS20PreWasmKeeper Create the IBC Transfer Stack from bottom to top:
// * SendPacket. Originates from the transferKeeper and goes up the stack:
// transferKeeper.SendPacket -> ibc_rate_limit.SendPacket -> ibc_hooks.SendPacket -> channel.SendPacket
// * RecvPacket, message that originates from core IBC and goes down to app, the flow is the other way
// channel.RecvPacket -> ibc_hooks.OnRecvPacket -> ibc_rate_limit.OnRecvPacket -> gmp.OnRecvPacket -> pfm.OnRecvPacket -> transfer.OnRecvPacket
//
// Note that the forward middleware is only integrated on the "receive" direction. It can be safely skipped when sending.
// Note also that the forward middleware is called "router", but we are using the name "pfm" (packet forward middleware) for clarity
// This may later be renamed upstream: https://github.com/ibc-apps/middleware/packet-forward-middleware/issues/10
//
// After this, the wasm keeper is required to be set on both
// app.Ics20WasmHooks AND app.RateLimitingICS4Wrapper
func (app *App) WireICS20PreWasmKeeper(
	appCodec codec.Codec,
) {
	// PFMKeeper must be created before TransferKeeper
	app.PFMKeeper = pfmkeeper.NewKeeper(
		appCodec,
		app.keys[pfmtypes.StoreKey],
		app.TransferKeeper.Keeper, // set later
		app.IBCKeeper.ChannelKeeper,
		app.FeeBurnerKeeper,
		&app.BankKeeper,
		app.IBCKeeper.ChannelKeeper,
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	wasmHooks := ibchooks.NewWasmHooks(nil, sdk.GetConfig().GetBech32AccountAddrPrefix()) // The contract keeper needs to be set later
	app.Ics20WasmHooks = &wasmHooks
	app.HooksICS4Wrapper = ibchooks.NewICS4Middleware(
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		&wasmHooks,
	)

	ibcratelimitKeeper := ibcratelimitkeeper.NewKeeper(appCodec, app.keys[ibcratelimittypes.ModuleName], authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String())
	// ChannelKeeper wrapper for rate limiting SendPacket(). The wasmKeeper needs to be added after it's created
	rateLimitingICS4Wrapper := ibcratelimit.NewICS4Middleware(
		app.HooksICS4Wrapper,
		&app.AccountKeeper,
		// wasm keeper we set later, right after wasmkeeper init. line 868
		nil,
		&app.BankKeeper,
		&ibcratelimitKeeper,
	)
	app.RateLimitingICS4Wrapper = &rateLimitingICS4Wrapper

	// Create Transfer Keepers
	app.TransferKeeper = wrapkeeper.NewKeeper(
		appCodec,
		app.keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.RateLimitingICS4Wrapper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		&app.BankKeeper,
		app.ScopedTransferKeeper,
		app.FeeKeeper,
		contractmanager.NewSudoLimitWrapper(app.ContractManagerKeeper, &app.WasmKeeper),
		authtypes.NewModuleAddress(adminmoduletypes.ModuleName).String(),
	)

	app.PFMKeeper.SetTransferKeeper(app.TransferKeeper.Keeper)

	// Packet Forward Middleware
	// Initialize packet forward middleware router
	var ibcStack ibcporttypes.IBCModule = packetforward.NewIBCMiddleware(
		transferSudo.NewIBCModule(
			app.TransferKeeper,
			contractmanager.NewSudoLimitWrapper(app.ContractManagerKeeper, &app.WasmKeeper),
		),
		app.PFMKeeper,
		0,
		pfmkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
		pfmkeeper.DefaultRefundTransferPacketTimeoutTimestamp,
	)

	ibcStack = gmpmiddleware.NewIBCMiddleware(ibcStack)
	// RateLimiting IBC Middleware
	rateLimitingTransferModule := ibcratelimit.NewIBCModule(ibcStack, app.RateLimitingICS4Wrapper)

	// Hooks Middleware
	hooksTransferModule := ibchooks.NewIBCMiddleware(&rateLimitingTransferModule, &app.HooksICS4Wrapper)
	app.TransferStack = &hooksTransferModule
}
