package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*authtypes.AccountI)(nil), &AbstractAccount{})
	registry.RegisterImplementations((*cryptotypes.PubKey)(nil), &NilPubKey{})

	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgRegisterAccount{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgUpdateParams{})

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
