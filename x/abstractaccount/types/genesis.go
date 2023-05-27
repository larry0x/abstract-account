package types

func DefaultGenesisState() *GenesisState {
	return &GenesisState{NextAccountId: 1}
}

func (GenesisState) Validate() error {
	return nil
}
