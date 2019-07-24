// Copyright (c) 2019 IoTeX Foundation
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package account

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-core/ioctl/cmd/alias"
	"github.com/iotexproject/iotex-core/ioctl/cmd/config"
	"github.com/iotexproject/iotex-core/ioctl/output"
	"github.com/iotexproject/iotex-core/ioctl/util"
)

// accountDeleteCmd represents the account delete command
var accountDeleteCmd = &cobra.Command{
	Use:   "delete [ALIAS|ADDRESS]",
	Short: "Delete an IoTeX account/address from wallet/config",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		if len(args) == 1 {
			err := accountDelete(args[0])
			return err
		}
		if config.ReadConfig.DefaultAccount.AddressOrAlias == "" {
			fmt.Println("Please specify a account to delete")
			return nil
		}
		err := accountDelete(config.ReadConfig.DefaultAccount.AddressOrAlias)
		return output.PrintError(err)
	},
}

func accountDelete(arg string) error {
	addr, err := util.GetAddress(arg)
	if err != nil {
		return output.NewError(output.AddressError, "failed to get address", err)
	}
	account, err := address.FromString(addr)
	if err != nil {
		return output.NewError(output.ConvertError, fmt.Sprintf("failed to convert string into address"),
			nil)
	}
	ks := keystore.NewKeyStore(config.ReadConfig.Wallet,
		keystore.StandardScryptN, keystore.StandardScryptP)
	for _, v := range ks.Accounts() {
		if bytes.Equal(account.Bytes(), v.Address.Bytes()) {
			var confirm string
			info := fmt.Sprintf("** This is an irreversible action!\n" +
				"Once an account is deleted, all the assets under this account may be lost!\n" +
				"Type 'YES' to continue, quit for anything else.")
			message := output.ConfirmationMessage{Info: info, Options: []string{"yes"}}
			fmt.Println(message.String())
			fmt.Scanf("%s", &confirm)
			if !strings.EqualFold(confirm, "yes") {
				output.PrintResult("quit")
				return nil
			}

			if err := os.Remove(v.URL.Path); err != nil {
				return output.NewError(output.WriteFileError, "failed to remove keystore file", err)
			}

			aliases := alias.GetAliasMap()
			delete(config.ReadConfig.Aliases, aliases[addr])
			out, err := yaml.Marshal(&config.ReadConfig)
			if err != nil {
				return output.NewError(output.SerializationError, "", err)
			}
			if err := ioutil.WriteFile(config.DefaultConfigFile, out, 0600); err != nil {
				return output.NewError(output.WriteFileError,
					fmt.Sprintf("Failed to write to config file %s.", config.DefaultConfigFile), err)
			}
			output.PrintResult(fmt.Sprintf("Account #%s has been deleted.", addr))
			return nil
		}
	}
	return output.NewError(output.ValidationError, fmt.Sprintf("account #%s not found", addr), nil)
}
