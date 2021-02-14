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
	"os/exec"
	"strconv"

	"github.com/Tatsuemon/anony-cli/lib"
	"github.com/Tatsuemon/anony/rpc"
	"github.com/jmoiron/sqlx"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func init() {
	rootCmd.AddCommand(newListAnonyURLsCmd())
}

type listAnonyURLsOpts struct {
	InActive bool
	All      bool
	Select   bool
}

func newListAnonyURLsCmd() *cobra.Command {
	opts := &listAnonyURLsOpts{}
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List your Anony URLs",
		Long:  "List your Anony URLs.\n By default, get Active Anony URLs.\nif you want to get in-active URLs, set 'inactive' flag.\nand if you want to get active and in-active URLs, set 'all' flag.",
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
			if err := listAnonyURLs(cmd, opts, db); err != nil {
				fmt.Println()
				return errors.Wrap(err, "failed to execute a command 'ls'\n")
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&opts.InActive, "inactive", "i", false, "get in-active URLs")
	cmd.Flags().BoolVarP(&opts.All, "all", "a", false, "get active & in-active URLs")
	cmd.Flags().BoolVarP(&opts.Select, "select", "s", false, "get URLs & select url you wanna open")
	return cmd
}

func listAnonyURLs(cmd *cobra.Command, opts *listAnonyURLsOpts, db *sqlx.DB) error {
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
	req := &rpc.ListAnonyURLsRequest{
		InActive: opts.InActive,
		All:      opts.All,
	}

	// Tokenのセット
	token, err := lib.GetToken(db, false)
	if err != nil {
		return err
	}
	md := metadata.Pairs("Authorization", fmt.Sprintf("bearer %s", token))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// API call
	res, err := cli.ListAnonyURLs(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to cli.ListAnonyURLs\n")
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"No.", "Original-URL", "Anony-URL", "active"})

	for i, v := range res.AnonyUrls {
		no := strconv.Itoa(i)
		active := strconv.FormatBool(v.IsActive)
		table.Append([]string{no, v.OriginalUrl, v.ShortUrl, active})
	}
	table.Render()
	if opts.Select {
		s := bufio.NewScanner(os.Stdin)
		var urlNoStr string
		fmt.Print("[Select URL No.]: ")
		s.Scan()
		urlNoStr = s.Text()
		urlNo, err := strconv.Atoi(urlNoStr)
		if err != nil {
			return err
		}
		if urlNo < 0 || urlNo >= len(res.AnonyUrls) {
			return fmt.Errorf("select No. is out of range")
		}
		// For Mac open
		err = exec.Command("open", res.AnonyUrls[urlNo].ShortUrl).Start()
		if err != nil {
			panic(err)
		}
	}
	return nil
}
