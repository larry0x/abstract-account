// NOTE TO MYSELF: do not confuse the following two interface types! They are
// identical but Go compiler doesn't consider them the same:
// - cosmos/cosmos-sdk/crypto/types.PubKey
// - cometbft/cometbft/crypto.PubKey

package types

import (
	"bytes"
	"errors"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/gogoproto/proto"
)

var (
	_ authtypes.AccountI = (*AbstractAccount)(nil)
	_ cryptotypes.PubKey = (*NilPubKey)(nil)
)

// ------------------------------ AbstractAccount ------------------------------

func NewAbstractAccount(address string, accountNum, seq uint64) *AbstractAccount {
	return &AbstractAccount{
		Address:       address,
		AccountNumber: accountNum,
		Sequence:      seq,
	}
}

func NewAbstractAccountFromAccount(acc authtypes.AccountI) *AbstractAccount {
	return NewAbstractAccount(acc.GetAddress().String(), acc.GetAccountNumber(), acc.GetSequence())
}

func (acc *AbstractAccount) GetAddress() sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(acc.Address)
	return addr
}

func (acc *AbstractAccount) SetAddress(addr sdk.AccAddress) error {
	if len(acc.Address) != 0 {
		return errors.New("cannot override AbstractAccount address")
	}

	acc.Address = addr.String()

	return nil
}

func (acc *AbstractAccount) GetPubKey() cryptotypes.PubKey {
	return NewNilPubKey(acc.GetAddress())
}

func (acc *AbstractAccount) SetPubKey(_ cryptotypes.PubKey) error {
	return errors.New("cannot set pubkey for AbstractAccount")
}

func (acc *AbstractAccount) GetAccountNumber() uint64 {
	return acc.AccountNumber
}

func (acc *AbstractAccount) SetAccountNumber(accNumber uint64) error {
	acc.AccountNumber = accNumber

	return nil
}

func (acc *AbstractAccount) GetSequence() uint64 {
	return acc.Sequence
}

func (acc *AbstractAccount) SetSequence(seq uint64) error {
	acc.Sequence = seq

	return nil
}

// --------------------------------- NilPubKey ---------------------------------

func NewNilPubKey(bz []byte) *NilPubKey {
	return &NilPubKey{AddressBytes: bz}
}

func (pk *NilPubKey) Address() cryptotypes.Address {
	return cryptotypes.Address(pk.AddressBytes)
}

func (pk *NilPubKey) Bytes() []byte {
	return nil
}

func (pk *NilPubKey) VerifySignature(_ []byte, _ []byte) bool {
	panic("NilPubKey.VerifySignature should never be invoked")
}

func (pk *NilPubKey) Equals(other cryptotypes.PubKey) bool {
	otherPk, ok := other.(*NilPubKey)
	if !ok {
		return false
	}

	return bytes.Equal(pk.AddressBytes, otherPk.AddressBytes)
}

func (pk *NilPubKey) Type() string {
	return "/" + proto.MessageName(pk)
}

func (pk *NilPubKey) String() string {
	return sdk.AccAddress(pk.AddressBytes).String()
}
