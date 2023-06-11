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

	return nil
}
