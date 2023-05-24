package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

const flagFunds = "funds"

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "abstract-account",
		Short:                      "Abstract account transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
		SilenceUsage:               true,
	}

	cmd.AddCommand(registerCmd())

	return cmd
}

func registerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register [code-id] [msg] --funds [coins,optional]",
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
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
		SilenceUsage: true,
	}

	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(flagFunds, "", "Coins to send to the account during instantiation")

	return cmd
}
