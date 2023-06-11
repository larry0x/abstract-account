package types

import "cosmossdk.io/errors"

var (
	ErrInvalidAllowedCodeIDs = errors.Register(ModuleName, 2, "AllowedCodeIDs must be empty if AllowAllCodeIDs is true")
	ErrNotAbstractAccount    = errors.Register(ModuleName, 3, "account is not an AbstractAccount")
	ErrNotAllowedCodeID      = errors.Register(ModuleName, 4, "not an allowed code ID")
	ErrNotBaseAccount        = errors.Register(ModuleName, 5, "account is not an authtypes.BaseAccount")
	ErrNotSingleSignautre    = errors.Register(ModuleName, 6, "signature is not a txsigning.SingleSignatureData")
	ErrParsingParams         = errors.Register(ModuleName, 7, "failed to marshal or unmarshal module params")
	ErrZeroMaxGas            = errors.Register(ModuleName, 8, "max gas cannot be zero")
)
