package keeper

import (
	"context"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

type msgServer struct {
	k Keeper
}

func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{k}
}

func (ms msgServer) RegisterAccount(goCtx context.Context, req *types.MsgRegisterAccount) (*types.MsgRegisterAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	senderAddr, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, err
	}

	contractAddr, data, err := ms.k.ck.Instantiate2(
		ctx,
		req.CodeID,
		senderAddr,
		senderAddr,
		req.Msg,
		fmt.Sprintf("%s/%d", types.ModuleName, ms.k.GetAndIncrementNextAccountID(ctx)),
		req.Funds,
		req.Salt,
		true, // I'm still not sure whether true is strictly better than false. Research needed
	)
	if err != nil {
		return nil, err
	}

	// set the contract's admin to itself
	if err = ms.k.ck.UpdateContractAdmin(ctx, contractAddr, senderAddr, contractAddr); err != nil {
		return nil, err
	}

	// the contract instantiation should have created a BaseAccount
	acc := ms.k.ak.GetAccount(ctx, contractAddr)
	if _, ok := acc.(*authtypes.BaseAccount); !ok {
		return nil, types.ErrNotBaseAccount
	}

	// we overwrite this BaseAccount with our AbstractAccount
	ms.k.ak.SetAccount(ctx, types.NewAbstractAccountFromAccount(acc))

	ms.k.Logger(ctx).Info(
		"abstract account registered",
		types.AttributeKeyCreator, req.Sender,
		types.AttributeKeyCodeID, req.CodeID,
		types.AttributeKeyContractAddr, contractAddr.String(),
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAccountRegistered,
			sdk.NewAttribute(types.AttributeKeyCreator, req.Sender),
			sdk.NewAttribute(types.AttributeKeyCodeID, strconv.FormatUint(req.CodeID, 10)),
			sdk.NewAttribute(types.AttributeKeyContractAddr, contractAddr.String()),
		),
	)

	return &types.MsgRegisterAccountResponse{Address: contractAddr.String(), Data: data}, nil
}
