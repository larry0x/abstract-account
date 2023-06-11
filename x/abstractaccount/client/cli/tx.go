package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

const (
	flagSalt  = "salt"
	flagFunds = "funds"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "abstract-account",
		Short:                      "Abstract account transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
		SilenceUsage:               true,
	}

	cmd.AddCommand(
		registerCmd(),
		migrateCmd(),
		updateParamsCmd(),
	)

	return cmd
}

func registerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register [code-id] [msg] --salt [string] --funds [coins,optional]",
		Short: "Register an abstract account",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			codeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			salt, err := cmd.Flags().GetString(flagSalt)
			if err != nil {
				return fmt.Errorf("salt: %s", err)
			}

			saltBytes := []byte(salt)
			if err = wasmtypes.ValidateSalt(saltBytes); err != nil {
				return fmt.Errorf("salt: %s", err)
			}

			amountStr, err := cmd.Flags().GetString(flagFunds)
			if err != nil {
				return fmt.Errorf("amount: %s", err)
			}

			amount, err := sdk.ParseCoinsNormalized(amountStr)
			if err != nil {
				return fmt.Errorf("amount: %s", err)
			}

			msg := &types.MsgRegisterAccount{
				Sender: clientCtx.GetFromAddress().String(),
				CodeID: codeID,
				Msg:    []byte(args[1]),
				Funds:  amount,
				Salt:   saltBytes,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
		SilenceUsage: true,
	}

	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(flagSalt, "", "Salt value used in determining account address")
	cmd.Flags().String(flagFunds, "", "Coins to send to the account during instantiation")

	return cmd
}

func migrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate [code-id] [msg]",
		Short: "Migrate an abstract account",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			codeID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := &types.MsgMigrateAccount{
				Sender: clientCtx.GetFromAddress().String(),
				CodeID: codeID,
				Msg:    []byte(args[1]),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
		SilenceUsage: true,
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func updateParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-params [json-encoded-params]",
		Short: "Update the module's parameters",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var params types.Params
			if err = json.Unmarshal([]byte(args[0]), &params); err != nil {
				return err
			}

			msg := &types.MsgUpdateParams{
				Sender: clientCtx.GetFromAddress().String(),
				Params: &params,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
		SilenceUsage: true,
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
