package types

const DefaultMaxGas = 2_000_000

func NewParams(maxGasBefore, maxGasAfter uint64) (*Params, error) {
	params := &Params{
		MaxGasBefore: maxGasBefore,
		MaxGasAfter:  maxGasAfter,
	}

	return params, params.Validate()
}

func DefaultParams() *Params {
	params, _ := NewParams(DefaultMaxGas, DefaultMaxGas)

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
