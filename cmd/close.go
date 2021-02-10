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

	"github.com/Tatsuemon/anony-cli/lib"
	"github.com/Tatsuemon/anony/rpc"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func init() {
	rootCmd.AddCommand(newCloseCmd())
}

type closeOpts struct {
	Original string
}

func newCloseCmd() *cobra.Command {
	opts := &closeOpts{}
	cmd := &cobra.Command{
		Use:   "close [oritinal url]",
		Short: "Deactivate your Anony URL",
		Long:  "Deactivate your Anony URL.",
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
			opts.Original = args[0]
			if err := close(cmd, opts, db); err != nil {
				fmt.Println()
				return errors.Wrap(err, "failed to execute a command 'close'\n")
			}
			return nil
		},
		Args: cobra.MinimumNArgs(1),
	}
	return cmd
}

func close(cmd *cobra.Command, opts *closeOpts, db *sqlx.DB) error {
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
	req := &rpc.UpdateAnonyURLStatusRequest{
		OriginalUrl: opts.Original,
		IsActive:    false,
	}

	// Tokenのセット
	token, err := lib.GetToken(db, false)
	if err != nil {
		return err
	}
	md := metadata.Pairs("Authorization", fmt.Sprintf("bearer %s", token))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// API call
	res, err := cli.UpdateAnonyURLStatus(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to cli.UpdateAnonyURLStatus\n")
	}

	// TODO(Tatsuemon): 出力の調整
	fmt.Println(res)

	return nil
}
