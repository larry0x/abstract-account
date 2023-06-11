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

// ------------------------------ RegisterAccount ------------------------------

func (ms msgServer) RegisterAccount(goCtx context.Context, req *types.MsgRegisterAccount) (*types.MsgRegisterAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params, err := ms.k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	// ensure that only allowed code IDs can be used to instantiate new accounts
	if !params.IsAllowedCodeID(req.CodeID) {
		return nil, types.ErrNotAllowedCodeID
	}

	senderAddr, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, err
	}

	contractAddr, data, err := ms.k.ck.Instantiate2(
		ctx,
		req.CodeID,
		senderAddr,
		// set module account as the admin
		//
		// previously we set the AbstractAccount itself as the admin. however, now
		// we want to enforce the code ID whitelist, we cannot allow the account to
		// migrate itself without the module's permission. therefore, now we set the
		// module account as admin, and provide a new MigrateAccount method.
		ms.k.ModuleAddress(),
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

// ------------------------------ MigrateAccount -------------------------------

func (ms msgServer) MigrateAccount(goCtx context.Context, req *types.MsgMigrateAccount) (*types.MsgMigrateAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params, err := ms.k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	// ensure that accounts can only be migrated to allowed code IDs
	if !params.IsAllowedCodeID(req.CodeID) {
		return nil, types.ErrNotAllowedCodeID
	}

	accAddr, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, err
	}

	// ensure that the account is indeed an AbstractAccount
	if _, ok := ms.k.ak.GetAccount(ctx, accAddr).(*types.AbstractAccount); !ok {
		return nil, types.ErrNotAbstractAccount
	}

	data, err := ms.k.ContractKeeper().Migrate(ctx, accAddr, ms.k.ModuleAddress(), req.CodeID, req.Msg)
	if err != nil {
		return nil, err
	}

	ms.k.Logger(ctx).Info(
		"account migrated",
		types.AttributeKeyContractAddr, accAddr.String(),
		types.AttributeKeyCodeID, req.CodeID,
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAccountMigrated,
			sdk.NewAttribute(types.AttributeKeyContractAddr, accAddr.String()),
			sdk.NewAttribute(types.AttributeKeyCodeID, strconv.FormatUint(req.CodeID, 10)),
		),
	)

	return &types.MsgMigrateAccountResponse{Data: data}, nil
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
