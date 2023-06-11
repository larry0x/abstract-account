package types

const DefaultMaxGas = 2_000_000

func NewParams(maxGasBefore, maxGasAfter uint64) *Params {
	return &Params{
		MaxGasBefore: maxGasBefore,
		MaxGasAfter:  maxGasAfter,
	}
}

func DefaultParams() *Params {
	return NewParams(DefaultMaxGas, DefaultMaxGas)
}

func (p *Params) Validate() error {
	if p.MaxGasBefore <= 0 {
		return ErrZeroMaxGas
	}

	if p.MaxGasAfter <= 0 {
		return ErrZeroMaxGas
	}

	if p.AllowAllCodeIDs && len(p.AllowedCodeIDs) > 0 {
		return ErrInvalidAllowedCodeIDs
	}

	return nil
}

func (p *Params) IsAllowedCodeID(codeID uint64) bool {
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
