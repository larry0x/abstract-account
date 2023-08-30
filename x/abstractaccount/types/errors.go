package types

import "cosmossdk.io/errors"

var (
	ErrMalformedAllowList = errors.Register(ModuleName, 2, "code ID allow list must contain non-zero, unique, and sorted code IDs")
	ErrNonEmptyAllowList  = errors.Register(ModuleName, 3, "code ID allow list must be empty when AllowAllCodeIDs is true")
	ErrNotAllowedCodeID   = errors.Register(ModuleName, 4, "not an allowed wasm code ID")
	ErrNotBaseAccount     = errors.Register(ModuleName, 5, "account is not an authtypes.BaseAccount")
	ErrNotSingleSignature = errors.Register(ModuleName, 6, "signature is not a txsigning.SingleSignatureData")
	ErrParsingParams      = errors.Register(ModuleName, 7, "failed to marshal or unmarshal module params")
	ErrZeroMaxGas         = errors.Register(ModuleName, 8, "max gas cannot be zero")
)
