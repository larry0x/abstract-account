package types

func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

func (GenesisState) Validate() error {
	return nil
}
