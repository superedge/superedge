/*
Copyright 2020 The SuperEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package token

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/superedge/superedge/pkg/util"
)

func NewTokenCMD() *cobra.Command {
	tokenCmd := &cobra.Command{
		Use:   "token",
		Short: "Manage bootstrap tokens",
	}

	createCmd := &cobra.Command{
		Use:                   "create [token]",
		DisableFlagsInUseLine: true,
		Short:                 "Create bootstrap tokens on the server",
		Long: `
			This command will create a bootstrap token for you.
			You can specify the usages for this token, the "time to live" and an optional human friendly description.

			The [token] is the actual token to write.
			This should be a securely generated random token of the form "[a-z0-9]{6}.[a-z0-9]{16}".
			If no [token] is given, kubeadm will generate a random token instead.
		`,
		RunE: func(tokenCmd *cobra.Command, args []string) error {
			//调用kubeadm创建token命令，创建一个bootstrap token
			createTokenCmd := fmt.Sprintf("kebeadm create token")
			outVal, _, err := util.RunLinuxCommand(createTokenCmd)
			if err != nil {
				return err
			}
			fmt.Println(outVal)
			return nil
		},
	}
	deleteCmd := &cobra.Command{
		Use:                   "delete [token-value] ...",
		DisableFlagsInUseLine: true,
		Short:                 "Delete bootstrap tokens on the server",
		Long: `
			This command will delete a list of bootstrap tokens for you.

			The [token-value] is the full Token of the form "[a-z0-9]{6}.[a-z0-9]{16}" or the
			Token ID of the form "[a-z0-9]{6}" to delete.
		`,
		RunE: func(tokenCmd *cobra.Command, args []string) error {
			if args[0] == "" {
				return errors.New("Please enter what you want to delete token")
			}

			deleteTokenCmd := fmt.Sprintf("kebeadm create delete %s", args[0])
			if _, _, err := util.RunLinuxCommand(deleteTokenCmd); err != nil {
				return err
			}
			return nil
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List bootstrap tokens on the server",
		RunE: func(tokenCmd *cobra.Command, args []string) error {
			tokenListCmd := fmt.Sprintf("kebeadm create delete %s", "test_token")
			outVal, _, err := util.RunLinuxCommand(tokenListCmd)
			if err != nil {
				return err
			}
			fmt.Println(outVal)
			return nil
		},
		Args: cobra.NoArgs,
	}

	tokenCmd.AddCommand(createCmd, deleteCmd, listCmd)
	return tokenCmd
}
