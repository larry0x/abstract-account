package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

var _ sdk.Msg = &MsgRegisterAccount{}

func (m *MsgRegisterAccount) ValidateBasic() error {
	msg := wasmtypes.MsgInstantiateContract{
		Sender: m.Sender,
		Admin:  m.Sender,
		CodeID: m.CodeID,
		Label:  AccountLabel(m.Sender, m.CodeID),
		Msg:    m.Msg,
		Funds:  m.Funds,
	}

	return msg.ValidateBasic()
}

func (m *MsgRegisterAccount) GetSigners() []sdk.AccAddress {
	// We've already asserted the sender address is valid in ValidateBasic, so we
	// can safety ignore the error here.
	senderAddr, _ := sdk.AccAddressFromBech32(m.Sender)

	return []sdk.AccAddress{senderAddr}
}
