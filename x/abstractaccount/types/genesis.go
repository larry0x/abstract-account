package types

func NewGenesisState(nextAccountID uint64) *GenesisState {
	return &GenesisState{
		NextAccountId: nextAccountID,
	}
}

func DefaultGenesisState() *GenesisState {
	return NewGenesisState(1)
}

func (GenesisState) Validate() error {
	return nil
}
