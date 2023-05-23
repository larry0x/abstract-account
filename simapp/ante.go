package simapp

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/larry0x/abstract-account/x/abstractaccount"
	abstractaccountkeeper "github.com/larry0x/abstract-account/x/abstractaccount/keeper"
)

// ----------------------------------- Ante ------------------------------------

type AnteHandlerOptions struct {
	ante.HandlerOptions

	AbstractAccountKeeper abstractaccountkeeper.Keeper
	WasmKeeper            wasm.Keeper
	WasmCfg               *wasmtypes.WasmConfig
	TXCounterStoreKey     storetypes.StoreKey
}

func NewAnteHandler(options AnteHandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("account keeper is required for AnteHandler")
	}

	if options.BankKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("bank keeper is required for AnteHandler")
	}

	if options.WasmCfg == nil {
		return nil, sdkerrors.ErrLogic.Wrap("wasm config is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, sdkerrors.ErrLogic.Wrap("sign mode handler is required for ante builder")
	}

	if options.TXCounterStoreKey == nil {
		return nil, sdkerrors.ErrLogic.Wrap("tx counter key is required for ante builder")
	}

	anteDecorators := []sdk.AnteDecorator{
		// outermost AnteDecorator. SetUpContext must be called first
		ante.NewSetUpContextDecorator(),
		// after setup context to enforce limits early
		wasmkeeper.NewLimitSimulationGasDecorator(options.WasmCfg.SimulationGasLimit),
		wasmkeeper.NewCountTXDecorator(options.TXCounterStoreKey),
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		// feegrant keeper set to nil since our simapp doesn't have a feegrant module
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, nil, options.TxFeeChecker),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		// BeforeTxKeeper replaces the default NewSigVerificationDecorator
		abstractaccount.NewBeforeTxDecorator(
			options.AbstractAccountKeeper,
			options.AccountKeeper,
			wasmkeeper.NewGovPermissionKeeper(options.WasmKeeper),
			options.SignModeHandler,
		),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}

// ----------------------------------- Post ------------------------------------

type PostHandlerOptions struct {
	posthandler.HandlerOptions

	AbstractAccountKeeper abstractaccountkeeper.Keeper
	AccountKeeper         ante.AccountKeeper
	WasmKeeper            wasm.Keeper
}

func NewPostHandler(options PostHandlerOptions) (sdk.PostHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("account keeper is required for AnteHandler")
	}

	postDecorators := []sdk.PostDecorator{
		abstractaccount.NewAfterTxDecorator(
			options.AbstractAccountKeeper,
			options.AccountKeeper,
			wasmkeeper.NewGovPermissionKeeper(options.WasmKeeper),
		),
	}

	return sdk.ChainPostDecorators(postDecorators...), nil
}
