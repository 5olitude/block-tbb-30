package main

import (
	"blocks/database"
	"blocks/node"
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var migrateCmd = func() *cobra.Command {
	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Migrates the blockchain db according to new bussiness rules",
		Run: func(cmd *cobra.Command, args []string) {

			state, err := database.NewStateFromDisk(getDataDirFromCmd(cmd))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer state.Close()
			pendingBlock := node.NewPendingBlock(
				database.Hash{},
				state.NextBlockNumber(),
				database.NewAccount("andrej"),
				[]database.Tx{
					database.NewTx("andrej", "andrej", 3, ""),
					database.NewTx("andrej", "babayaga", 2000, ""),
					database.NewTx("babayaga", "andrej", 1, ""),
					database.NewTx("babayaga", "caesar", 1000, ""),
					database.NewTx("babayaga", "andrej", 50, ""),
				},
			)
			_, err = node.Mine(context.Background(), pendingBlock)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
	addDefaultRequiredFlags(migrateCmd)
	return migrateCmd
}
