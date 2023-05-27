package simapp

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
)

type GenesisState map[string]json.RawMessage

func DefaultGenesisState(cdc codec.JSONCodec) GenesisState {
	return ModuleBasics.DefaultGenesis(cdc)
}
