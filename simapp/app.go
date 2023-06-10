package simapp

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cast"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tmos "github.com/cometbft/cometbft/libs/os"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/std"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensus "github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/larry0x/simapp/x/poa"
	poatypes "github.com/larry0x/simapp/x/poa/types"

	"github.com/larry0x/abstract-account/x/abstractaccount"
	abstractaccountkeeper "github.com/larry0x/abstract-account/x/abstractaccount/keeper"
	abstractaccounttypes "github.com/larry0x/abstract-account/x/abstractaccount/types"
)

const (
	appName = "SimApp"

	// A random account I created to serve as the authority for modules, since
	// this simapp doesn't have a gov module.
	//
	// The seed phrase is:
	//
	// crumble soon   hockey  pigeon  border   health
	// human   cotton romance fork    mountain rapid
	// scan    swarm  basic   subject tornado  genius
	// parade  stone  coyote  pluck   journey  fatal
	authority = "cosmos1tqr9a9m9nk0c22uq2c2slundmqhtnrnhwks7x0"
)

var (
	DefaultNodeHome string

	ModuleBasics = module.NewBasicManager(
		abstractaccount.AppModuleBasic{},
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		consensus.AppModuleBasic{},
		poa.AppModuleBasic{},
		wasm.AppModuleBasic{},
	)

	maccPerms = map[string][]string{
		authtypes.FeeCollectorName: nil,
		wasm.ModuleName:            {authtypes.Burner},
	}
)

var (
	_ runtime.AppI            = (*SimApp)(nil)
	_ servertypes.Application = (*SimApp)(nil)
)

type SimApp struct {
	*baseapp.BaseApp

	amino             *codec.LegacyAmino
	cdc               codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry types.InterfaceRegistry

	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	AbstractAccountKeeper abstractaccountkeeper.Keeper
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	WasmKeeper            wasm.Keeper

	ModuleManager *module.Manager
	configurator  module.Configurator
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, ".simapp")
}

func NewSimApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	wasmOpts []wasm.Option,
	baseAppOptions ...func(*baseapp.BaseApp),
) *SimApp {
	encCfg := MakeEncodingConfig()

	bApp := baseapp.NewBaseApp(
		appName,
		logger,
		db,
		encCfg.TxConfig.TxDecoder(),
		baseAppOptions...,
	)

	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(encCfg.InterfaceRegistry)
	bApp.SetTxEncoder(encCfg.TxConfig.TxEncoder())

	keys := sdk.NewKVStoreKeys(
		abstractaccounttypes.StoreKey,
		authtypes.StoreKey,
		banktypes.StoreKey,
		consensusparamtypes.StoreKey,
		wasm.StoreKey,
	)
	tkeys := sdk.NewTransientStoreKeys()
	memKeys := sdk.NewMemoryStoreKeys()

	app := &SimApp{
		BaseApp:           bApp,
		amino:             encCfg.Amino,
		cdc:               encCfg.Codec,
		txConfig:          encCfg.TxConfig,
		interfaceRegistry: encCfg.InterfaceRegistry,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		app.cdc,
		keys[consensusparamtypes.StoreKey],
		authority,
	)
	app.SetParamStore(&app.ConsensusParamsKeeper)

	app.AccountKeeper = authkeeper.NewAccountKeeper(
		app.cdc,
		keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		maccPerms,
		sdk.Bech32MainPrefix,
		authority,
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		app.cdc,
		keys[banktypes.StoreKey],
		app.AccountKeeper,
		blockedAddresses(),
		authority,
	)

	wasmDir, wasmCfg, wasmCapabilities := wasmParams(appOpts)
	app.WasmKeeper = wasm.NewKeeper(
		app.cdc,
		keys[wasm.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		wasmCfg,
		wasmCapabilities,
		authority,
		wasmOpts...,
	)

	app.AbstractAccountKeeper = abstractaccountkeeper.NewKeeper(
		app.cdc,
		keys[abstractaccounttypes.StoreKey],
		app.AccountKeeper,
		// we don't really need this strong permission (we don't need to store code
		// or modify code access config) but wasm module doesn't seem to allow us
		// to create our own authorization policy
		wasmkeeper.NewGovPermissionKeeper(app.WasmKeeper),
		authority,
	)

	app.ModuleManager = module.NewManager(
		abstractaccount.NewAppModule(app.AbstractAccountKeeper),
		auth.NewAppModule(app.cdc, app.AccountKeeper, authsims.RandomGenesisAccounts, nil),
		bank.NewAppModule(app.cdc, app.BankKeeper, app.AccountKeeper, nil),
		consensus.NewAppModule(app.cdc, app.ConsensusParamsKeeper),
		poa.NewAppModule(),
		wasm.NewAppModule(app.cdc, &app.WasmKeeper, nil, app.AccountKeeper, app.BankKeeper, app.MsgServiceRouter(), nil),
	)

	app.ModuleManager.SetOrderBeginBlockers(
		poatypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		consensusparamtypes.ModuleName,
		wasm.ModuleName,
		abstractaccounttypes.ModuleName,
	)

	app.ModuleManager.SetOrderEndBlockers(
		poatypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		consensusparamtypes.ModuleName,
		wasm.ModuleName,
		abstractaccounttypes.ModuleName,
	)

	genesisModuleOrder := []string{
		authtypes.ModuleName,
		banktypes.ModuleName,
		poatypes.ModuleName,
		consensusparamtypes.ModuleName,
		wasm.ModuleName,
		abstractaccounttypes.ModuleName,
	}
	app.ModuleManager.SetOrderInitGenesis(genesisModuleOrder...)
	app.ModuleManager.SetOrderExportGenesis(genesisModuleOrder...)

	app.configurator = module.NewConfigurator(app.cdc, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.ModuleManager.RegisterServices(app.configurator)

	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	app.setAnteHandler(encCfg.TxConfig, wasmCfg, keys[wasm.StoreKey])
	app.setPostHandler()

	if manager := app.SnapshotManager(); manager != nil {
		if err := manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.WasmKeeper),
		); err != nil {
			panic(fmt.Errorf("failed to register snapshot extension: %s", err))
		}
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			logger.Error("error on loading last version", "err", err)
			os.Exit(1)
		}

		ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})
		if err := app.WasmKeeper.InitializePinnedCodes(ctx); err != nil {
			tmos.Exit(fmt.Sprintf("failed initialize pinned codes %s", err))
		}
	}

	return app
}

func (app *SimApp) setAnteHandler(txCfg client.TxConfig, wasmCfg wasmtypes.WasmConfig, txCounterStoreKey storetypes.StoreKey) {
	anteHandler, err := NewAnteHandler(
		AnteHandlerOptions{
			HandlerOptions: ante.HandlerOptions{
				AccountKeeper:   app.AccountKeeper,
				BankKeeper:      app.BankKeeper,
				SignModeHandler: txCfg.SignModeHandler(),
				SigGasConsumer:  abstractaccount.SigVerificationGasConsumer,
			},
			WasmCfg:               &wasmCfg,
			TXCounterStoreKey:     txCounterStoreKey,
			AbstractAccountKeeper: app.AbstractAccountKeeper,
		},
	)
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)
}

func (app *SimApp) setPostHandler() {
	postHandler, err := NewPostHandler(
		PostHandlerOptions{
			HandlerOptions:        posthandler.HandlerOptions{},
			AccountKeeper:         app.AccountKeeper,
			AbstractAccountKeeper: app.AbstractAccountKeeper,
		},
	)
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)
}

// ------------------------------- runtime.AppI --------------------------------

func (app *SimApp) Name() string {
	return app.BaseApp.Name()
}

func (app *SimApp) LegacyAmino() *codec.LegacyAmino {
	return app.amino
}

func (app *SimApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	return app.ModuleManager.InitGenesis(ctx, app.cdc, genesisState)
}

func (app *SimApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.ModuleManager.BeginBlock(ctx, req)
}

func (app *SimApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.ModuleManager.EndBlock(ctx, req)
}

func (app *SimApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

func (app *SimApp) ExportAppStateAndValidators(_ bool, _ []string, _ []string) (servertypes.ExportedApp, error) {
	panic("UNIMPLEMENTED")
}

func (app *SimApp) SimulationManager() *module.SimulationManager {
	panic("UNIMPLEMENTED")
}

// -------------------------- servertypes.Application --------------------------

func (app *SimApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx

	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

func (app *SimApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(
		app.BaseApp.GRPCQueryRouter(),
		clientCtx,
		app.BaseApp.Simulate,
		app.interfaceRegistry,
	)
}

func (app *SimApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

func (app *SimApp) RegisterNodeService(clientCtx client.Context) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}

// ----------------------------------- Misc ------------------------------------

func (app *SimApp) Codec() codec.Codec {
	return app.cdc
}

func (app *SimApp) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

func (app *SimApp) TxConfig() client.TxConfig {
	return app.txConfig
}

func MakeEncodingConfig() EncodingConfig {
	encCfg := MakeTestEncodingConfig()

	std.RegisterLegacyAminoCodec(encCfg.Amino)
	std.RegisterInterfaces(encCfg.InterfaceRegistry)

	ModuleBasics.RegisterLegacyAminoCodec(encCfg.Amino)
	ModuleBasics.RegisterInterfaces(encCfg.InterfaceRegistry)

	return encCfg
}

func blockedAddresses() map[string]bool {
	modAccAddrs := make(map[string]bool)

	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

func wasmParams(appOpts servertypes.AppOptions) (string, wasmtypes.WasmConfig, string) {
	// dir
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))
	wasmDir := filepath.Join(homePath, "wasm")

	// config
	wasmCfg, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}

	// capabilities
	wasmCapabilities := "iterator,staking,stargate,cosmwasm_1_1,cosmwasm_1_2"

	return wasmDir, wasmCfg, wasmCapabilities
}
