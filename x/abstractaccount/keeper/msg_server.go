package keeper

import (
	"context"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

type msgServer struct {
	k Keeper
}

func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{k}
}

// ------------------------------- UpdateParams --------------------------------

func (ms msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req.Sender != ms.k.authority {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("sender is not authority: expect %s, found %s", ms.k.authority, req.Sender)
	}

	if err := req.Params.Validate(); err != nil {
		return nil, err
	}

	if err := ms.k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// ------------------------------ RegisterAccount ------------------------------

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
		// we set fix_msg to false because there simply isn't any good reason
		// otherwise, given that we already have full control over the address by
		// providing a salt. read more:
		// https://medium.com/cosmwasm/dev-note-3-limitations-of-instantiate2-and-how-to-deal-with-them-a3f946874230
		false,
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
		"account registered",
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
