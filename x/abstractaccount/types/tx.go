package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgRegisterAccount{}
)

// ------------------------------- UpdateParams --------------------------------

func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrap("invalid sender address")
	}

	return m.Params.Validate()
}

func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	// We've already asserted the sender address is valid in ValidateBasic, so we
	// can safety ignore the error here.
	senderAddr, _ := sdk.AccAddressFromBech32(m.Sender)

	return []sdk.AccAddress{senderAddr}
}

// ------------------------------ RegisterAccount ------------------------------

func (m *MsgRegisterAccount) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrap("invalid sender address")
	}

	if m.CodeID == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("code id cannot be zero")
	}

	if err := m.Msg.ValidateBasic(); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid init msg: %s", err.Error())
	}

	if !m.Funds.IsValid() {
		return sdkerrors.ErrInvalidCoins
	}

	return wasmtypes.ValidateSalt(m.Salt)
}

func (m *MsgRegisterAccount) GetSigners() []sdk.AccAddress {
	senderAddr, _ := sdk.AccAddressFromBech32(m.Sender)

	return []sdk.AccAddress{senderAddr}
}
