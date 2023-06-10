package types

const (
	ModuleName = "abstractaccount"
	StoreKey   = ModuleName
)

var (
	// Module parameters
	KeyParams = []byte{0x00}

	// We give each AbstractAccount an ID - The first to be registered gets an ID
	// of 1, the second gets 2, so on.
	//
	// This is used to determine the account smart contract's label. Specifically,
	// the label is `abstractaccount/{id}`.
	//
	// The label is only for identifying contracts; it doesn't impact the actual
	// working of these contract any way. However, I just like everything cleanly
	// and uniquely labeled.
	KeyNextAccountID = []byte{0x01}

	// In the AnteHandler, if the tx only has one sender and this sender is an
	// AbstractAccount, we store its address here. This way, in the PostHandler,
	// we know whether to call the after_tx method.
	KeySignerAddress = []byte{0x02}
)
