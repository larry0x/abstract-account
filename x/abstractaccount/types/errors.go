package types

import "cosmossdk.io/errors"

var (
	ErrNotAllowedCodeID   = errors.Register(ModuleName, 2, "not an allowed wasm code ID")
	ErrNotBaseAccount     = errors.Register(ModuleName, 3, "account is not an authtypes.BaseAccount")
	ErrNotSingleSignautre = errors.Register(ModuleName, 4, "signature is not a txsigning.SingleSignatureData")
	ErrParsingParams      = errors.Register(ModuleName, 5, "failed to marshal or unmarshal module params")
	ErrZeroMaxGas         = errors.Register(ModuleName, 6, "max gas cannot be zero")
)
