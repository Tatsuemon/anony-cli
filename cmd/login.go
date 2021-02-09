/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/Tatsuemon/anony/rpc"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
	"google.golang.org/grpc"
)

func init() {
	rootCmd.AddCommand(newLogInCmd())
}

type logInOpts struct {
	NameOrEmail string
	Password    string
}

func newLogInCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in",
		Long:  `Log in to Anony service in order to use Anony CLI commands.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// fmt.Println(config)
			s := bufio.NewScanner(os.Stdin)
			var nameOrEmail, password string
			fmt.Print("[Account Name or Email]: ")
			s.Scan()
			nameOrEmail = s.Text()
			fmt.Print("[Password]: ")
			pwd, err := terminal.ReadPassword(syscall.Stdin)
			password = string(pwd)
			if err != nil {
				fmt.Println(err)
				return nil
			}
			opts := &logInOpts{
				NameOrEmail: nameOrEmail,
				Password:    password,
			}
			if err := logInUser(cmd, opts); err != nil {
				fmt.Println()
				return errors.Wrap(err, "failed to execute a command 'login'\n")
			}
			return nil
		},
	}
	return cmd
}

func logInUser(cmd *cobra.Command, opts *logInOpts) error {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())

	if err != nil {
		return errors.Wrap(err, "failed to establish connection\n")
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("failed to conn.Close(): \n%v", err)
		}
	}()

	cli := rpc.NewUserServiceClient(conn)

	req := &rpc.LogInUserRequest{
		NameOrEmail: opts.NameOrEmail,
		Password:    opts.Password,
	}

	res, err := cli.LogInUser(context.Background(), req)
	// TODO(Tatsuemon): パスワードとName, Emailが違う時はError表示はしない
	if err != nil {
		return errors.Wrap(err, "failed to cli.LogInUser\n")
	}
	fmt.Printf("\n\nHi %s!! You've successfully authenticated.\n", res.GetUser().GetName())
	fmt.Println("Welcome to Anonny!!")

	// Tokenの設定
	file, err := os.OpenFile(viper.ConfigFileUsed(), os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	fmt.Fprintf(file, "Token: \"%s\"\n", res.GetToken())

	return nil
}
