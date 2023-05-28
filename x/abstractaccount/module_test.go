// there isn't anything to test for in module.go
// instead this file contains helper functions to be used for testing ante.go

package abstractaccount_test

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/larry0x/abstract-account/simapp"
	"github.com/larry0x/abstract-account/x/abstractaccount/keeper"
	"github.com/larry0x/abstract-account/x/abstractaccount/testdata"
	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

func storeCodeAndRegisterAccount(
	ctx sdk.Context, app *simapp.SimApp, senderAddr sdk.AccAddress,
	bytecode []byte, msg interface{}, funds sdk.Coins,
) (codeID uint64, contractAddr sdk.Address, err error) {
	k := app.AbstractAccountKeeper
	msgServer := keeper.NewMsgServerImpl(k)

	// store code
	codeID, _, err = k.ContractKeeper().Create(ctx, senderAddr, testdata.AccountWasm, nil)
	if err != nil {
		return 0, nil, err
	}

	// prepare the contract instantiate msg
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return 0, nil, err
	}

	// register the account
	res, err := msgServer.RegisterAccount(ctx, &types.MsgRegisterAccount{
		Sender: senderAddr.String(),
		CodeID: codeID,
		Msg:    msgBytes,
		Funds:  funds,
	})
	if err != nil {
		return 0, nil, err
	}

	contractAddr, err = sdk.AccAddressFromBech32(res.Address)
	if err != nil {
		return 0, nil, err
	}

	return codeID, contractAddr, nil
}
