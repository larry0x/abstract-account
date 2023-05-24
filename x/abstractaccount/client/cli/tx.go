package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

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

	cmd.AddCommand(
		registerCmd(),
		signCmd(),
	)

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

func signCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign [file] --from [key] --output-document [file]",
		Short: "Sign a transaction for an abstract account",
		Long: `A technical note on why we can't use the "simd tx sign" command: that
command does two things that doesn't make sense for AbstractAccounts (AAs):

- It populates the tx with a secp256k1 pubkey, while AAs don't store pubkeys
	at the sdk level;
- it asserts that the pubkey derives an address that matches the tx's sender,
	which isn't the case for AAs for which the addresses are derived by the wasm
	module using a different set of rules.

This command on the other hand, populates the tx with a special NilPubKey type,
while also skips the address check.
`,
		Args:    cobra.ExactArgs(1),
		PreRunE: assertOnlyOnlineMode,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			stdTx, err := authclient.ReadTxFromFile(clientCtx, args[0])
			if err != nil {
				return err
			}

			signerAcc, err := getSignerOfTx(clientCtx, stdTx)
			if err != nil {
				return err
			}

			signerData := authsigning.SignerData{
				Address:       signerAcc.GetAddress().String(),
				ChainID:       clientCtx.ChainID,
				AccountNumber: signerAcc.GetAccountNumber(),
				Sequence:      signerAcc.GetSequence(),
				PubKey:        signerAcc.GetPubKey(),
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			txBuilder, err := clientCtx.TxConfig.WrapTxBuilder(stdTx)
			if err != nil {
				return err
			}

			sigData := txsigning.SingleSignatureData{
				SignMode:  txf.SignMode(),
				Signature: nil,
			}

			sig := txsigning.SignatureV2{
				PubKey:   signerAcc.GetPubKey(),
				Data:     &sigData,
				Sequence: txf.Sequence(),
			}

			if err := txBuilder.SetSignatures(sig); err != nil {
				return err
			}

			signBytes, err := clientCtx.TxConfig.SignModeHandler().GetSignBytes(txf.SignMode(), signerData, txBuilder.GetTx())
			if err != nil {
				return err
			}

			from, err := cmd.Flags().GetString(flags.FlagFrom)
			if err != nil {
				return err
			}

			sigBytes, _, err := txf.Keybase().Sign(from, signBytes)
			if err != nil {
				return err
			}

			sigData = txsigning.SingleSignatureData{
				SignMode:  txf.SignMode(),
				Signature: sigBytes,
			}

			sig = txsigning.SignatureV2{
				PubKey:   signerAcc.GetPubKey(),
				Data:     &sigData,
				Sequence: txf.Sequence(),
			}

			if err := txBuilder.SetSignatures(sig); err != nil {
				return err
			}

			if err := txf.PreprocessTx(from, txBuilder); err != nil {
				return err
			}

			json, err := clientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
			if err != nil {
				return err
			}

			outputDoc, _ := cmd.Flags().GetString(flags.FlagOutputDocument)
			fp, err := os.OpenFile(outputDoc, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
			if err != nil {
				return err
			}

			cmd.SetOut(fp)
			cmd.Printf("%s\n", json)

			fp.Close()

			fmt.Printf("Signed tx written to %s\n", outputDoc)

			return nil
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(flags.FlagOutputDocument, "", "The document will be written to the given file instead of STDOUT")

	cmd.MarkFlagRequired(flags.FlagFrom)
	cmd.MarkFlagRequired(flags.FlagOutputDocument)

	return cmd
}

func getSignerOfTx(clientCtx client.Context, stdTx sdk.Tx) (*types.AbstractAccount, error) {
	var signerAddr sdk.AccAddress = nil
	for i, msg := range stdTx.GetMsgs() {
		signers := msg.GetSigners()
		if len(signers) != 1 {
			return nil, fmt.Errorf("msg %d has more than one signers", i)
		}

		if signerAddr != nil && bytes.Equal(signerAddr, signers[0]) {
			return nil, errors.New("tx has more than one signers")
		}

		signerAddr = signers[0]
	}

	signerAcc, err := clientCtx.AccountRetriever.GetAccount(clientCtx, signerAddr)
	if err != nil {
		return nil, err
	}

	absAcc, ok := signerAcc.(*types.AbstractAccount)
	if !ok {
		return nil, fmt.Errorf("account %s is not an AbstractAccount", signerAcc.GetAddress().String())
	}

	return absAcc, nil
}

func assertOnlyOnlineMode(cmd *cobra.Command, _ []string) error {
	if offline, _ := cmd.Flags().GetBool(flags.FlagOffline); offline {
		return errors.New("use online mode")
	}

	if accountNum, _ := cmd.Flags().GetUint64(flags.FlagAccountNumber); accountNum != 0 {
		return errors.New("use online mode, don't specify the account number")
	}

	if sequence, _ := cmd.Flags().GetUint64(flags.FlagSequence); sequence != 0 {
		return errors.New("use online mode, don't specify the sequence")
	}

	return nil
}
