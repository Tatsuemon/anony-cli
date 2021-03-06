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
	"context"
	"fmt"
	"log"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/Tatsuemon/anony-cli/lib"
	"github.com/Tatsuemon/anony/rpc"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func init() {
	rootCmd.AddCommand(newInfoCmd())
}

type infoLOpts struct{}

func newInfoCmd() *cobra.Command {
	opts := &infoLOpts{}
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show your info",
		Long:  "Show your infomation",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !lib.ExistDB(lib.GetDBPath()) {
				return fmt.Errorf("failed to login, please initialize anony")
			}
			// DB Conn
			db, err := sqlx.Open("sqlite3", lib.GetDBPath())
			if err != nil {
				log.Fatal(err)
			}
			defer db.Close()
			if err := info(cmd, opts, db); err != nil {
				fmt.Println()
				return errors.Wrap(err, "failed to execute a command 'info'\n")
			}
			return nil
		},
	}
	return cmd
}

func info(cmd *cobra.Command, opts *infoLOpts, db *sqlx.DB) error {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())

	if err != nil {
		return errors.Wrap(err, "failed to establish connection\n")
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("failed to conn.Close(): \n%v", err)
		}
	}()

	cli := rpc.NewAnonyServiceClient(conn)
	req := &emptypb.Empty{}

	// Tokenのセット
	token, err := lib.GetToken(db, false)
	if err != nil {
		return err
	}
	md := metadata.Pairs("Authorization", fmt.Sprintf("bearer %s", token))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// API call
	res, err := cli.CountAnonyURLs(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to cli.CountAnonyURLs\n")
	}

	fmt.Printf("Name: %s\n", res.Name)
	fmt.Printf("Email: %s\n", res.Email)
	fmt.Printf("URLs(Active): %v(%v)\n", res.CountAll, res.CountActive)
	return nil
}
