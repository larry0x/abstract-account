package types

func NewGenesisState(nextAccountID uint64, params *Params) *GenesisState {
	return &GenesisState{
		NextAccountId: nextAccountID,
		Params:        params,
	}
}

func DefaultGenesisState() *GenesisState {
	return NewGenesisState(1, DefaultParams())
}

func (gs *GenesisState) Validate() error {
	return gs.Params.Validate()
}
