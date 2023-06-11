package types

import "cosmossdk.io/errors"

var (
	ErrNotBaseAccount     = errors.Register(ModuleName, 2, "account is not an authtypes.BaseAccount")
	ErrNotSingleSignautre = errors.Register(ModuleName, 3, "signature is not a txsigning.SingleSignatureData")
	ErrParsingParams      = errors.Register(ModuleName, 4, "failed to marshal or unmarshal module params")
	ErrZeroMaxGas         = errors.Register(ModuleName, 5, "max gas cannot be zero")
)
