package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var amino = codec.NewLegacyAmino()

// NOTE: This is required for the GetSignBytes function
func init() {
	RegisterLegacyAminoCodec(amino)

	sdk.RegisterLegacyAminoCodec(amino)
	// cryptocodec.RegisterCrypto(amino)
	// codec.RegisterEvidences(amino)

	// Register all Amino interfaces and concrete types on the authz Amino codec
	// so that this can later be used to properly serialize MsgGrant and MsgExec
	// instances.
	RegisterLegacyAminoCodec(legacy.Cdc)

	amino.Seal()
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "abstract-account/MsgUpdateParams")

	cdc.RegisterConcrete(&NilPubKey{}, "abstract-account/NilPubKey", nil)
	cdc.RegisterConcrete(&Params{}, "abstract-account/Params", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.AccountI)(nil), &AbstractAccount{})
	registry.RegisterImplementations(
		(*authtypes.GenesisAccount)(nil),
		&AbstractAccount{},
	)
	registry.RegisterImplementations((*cryptotypes.PubKey)(nil), &NilPubKey{})

	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgRegisterAccount{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgUpdateParams{})

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
