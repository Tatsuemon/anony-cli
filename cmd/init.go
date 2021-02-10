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
	"log"
	"os"
	"syscall"
	"unicode/utf8"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/Tatsuemon/anony-cli/lib"
	"github.com/Tatsuemon/anony/rpc"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"google.golang.org/grpc"
)

func init() {
	rootCmd.AddCommand(newInitCmd())
}

type initOpts struct {
	Name            string
	Email           string
	Password        string
	ConfirmPassword string
}

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize user information",
		Long:  `Create user account & log in Anony in order to use Annony CLI commands`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := bufio.NewScanner(os.Stdin)
			var name, email, password, confirmPassword string
			fmt.Println("Hello!!!\nWelcome to Anony!!\n\nYou need to input your nickname and email.")
			fmt.Print("[Account Name]: ")
			s.Scan()
			name = s.Text()
			fmt.Print("[Email]: ")
			s.Scan()
			email = s.Text()
			fmt.Println()
			fmt.Println("Please enter a password of at least 6 characters.")
			for i := 0; i < 3; i++ {
				fmt.Print("[Password]: ")
				pwd, err := terminal.ReadPassword(syscall.Stdin)
				password = string(pwd)
				if err != nil {
					fmt.Println(err)
					return nil
				}
				fmt.Println()
				if utf8.RuneCountInString(password) < 6 {
					fmt.Println("** Password must be at least 6 characters **")
					if i == 2 {
						fmt.Println("** Please start over from the beginning **")
						return nil
					}
				} else {
					break
				}
			}
			for i := 0; i < 3; i++ {
				fmt.Print("[Confirm Password]: ")
				pwd, err := terminal.ReadPassword(syscall.Stdin)
				confirmPassword = string(pwd)
				if err != nil {
					fmt.Println(err)
					return nil
				}
				fmt.Println()
				if password != confirmPassword {
					fmt.Println("** Please enter the same Password **")
					if i == 2 {
						fmt.Println("** Please start over from the beginning **")
						return nil
					}
				} else {
					break
				}
			}
			// DB Conn
			db, err := sqlx.Open("sqlite3", lib.GetDBPath())
			if err != nil {
				log.Fatal(err)
			}
			defer db.Close()

			opts := &initOpts{
				Name:            name,
				Email:           email,
				Password:        password,
				ConfirmPassword: confirmPassword,
			}
			err = initUser(cmd, opts, db)
			if err != nil {
				fmt.Println()
				return errors.Wrap(err, "failed to execute a command 'init'\n")
			}

			return nil
		},
	}
	return cmd
}

func initUser(cmd *cobra.Command, opts *initOpts, db *sqlx.DB) error {
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

	req := &rpc.CreateUserRequest{
		User: &rpc.UserBase{
			Name:  opts.Name,
			Email: opts.Email,
		},
		Password:        opts.Password,
		ConfirmPassword: opts.ConfirmPassword,
	}

	res, err := cli.CreateUser(context.Background(), req)
	// TODO(Tatsuemon): already existsはErrorではないため, Error表示しないようにする
	if err != nil {
		return errors.Wrap(err, "failed to cli.CreateUser\n")
	}
	// DBの作成
	if err = lib.CreateDBTable(db); err != nil {
		return err
	}
	// Tokenのセット
	if err = lib.InsertToken(db, res.GetToken(), false); err != nil {
		return err
	}

	fmt.Printf("\nHi %s!! You've successfully registered user.\n", res.GetUser().GetName())
	return nil
}
