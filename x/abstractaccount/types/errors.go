package types

import "cosmossdk.io/errors"

var ErrNotSingleSignautre = errors.Register(ModuleName, 2, "signature is not a txsigning.SingleSignatureData")
