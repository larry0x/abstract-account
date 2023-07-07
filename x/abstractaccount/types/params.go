package types

const DefaultMaxGas = 2_000_000

func NewParams(allowAllCodeIDs bool, allowedCodeIDs []uint64, maxGasBefore, maxGasAfter uint64) (*Params, error) {
	params := &Params{
		AllowAllCodeIDs: allowAllCodeIDs,
		AllowedCodeIDs:  allowedCodeIDs,
		MaxGasBefore:    maxGasBefore,
		MaxGasAfter:     maxGasAfter,
	}

	return params, params.Validate()
}

func DefaultParams() *Params {
	params, _ := NewParams(true, []uint64{}, DefaultMaxGas, DefaultMaxGas)

	return params
}

func (p *Params) Validate() error {
	if p.MaxGasBefore <= 0 {
		return ErrZeroMaxGas
	}

	if p.MaxGasAfter <= 0 {
		return ErrZeroMaxGas
	}

	return nil
}

// IsAllowed returns whether a code ID is allowed to be used to register
// AbstractAccounts.
func (p *Params) IsAllowed(codeID uint64) bool {
	if p.AllowAllCodeIDs {
		return true
	}

	for _, allowedCodeID := range p.AllowedCodeIDs {
		if codeID == allowedCodeID {
			return true
		}
	}

	return false
}
