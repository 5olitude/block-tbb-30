package main

import (
	"blocks/node"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func runCmd() *cobra.Command {
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Launches the TBB node and its HTTP API .",
		Run: func(cmd *cobra.Command, args []string) {
			port, _ := cmd.Flags().GetUint64(flagPort)
			fmt.Println("launching TBB node and its HTTP API ...")
			bootstrap := node.NewPeerNode(
				"127.0.0.1",
				8080,
				true,
				true,
			)
			n := node.New(getDataDirFromCmd(cmd), port, bootstrap)
			err := n.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	addDefaultRequiredFlags(runCmd)
	runCmd.Flags().Uint64(flagPort, node.DefaultHTTPort, "exposed http port for communication with peers")
	return runCmd
}
