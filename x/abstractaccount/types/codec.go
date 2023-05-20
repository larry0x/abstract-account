package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*authtypes.AccountI)(nil), &AbstractAccount{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgRegisterAccount{})
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
