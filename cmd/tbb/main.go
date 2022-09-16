package main

import (
	"blocks/fs"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

const flagDataDir = "datadir"
const flagPort = "port"
const flagIP = "ip"

func main() {
	var tbbCmd = &cobra.Command{
		Use:   "tbb",
		Short: "The Blockchain Bar CLI",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	tbbCmd.AddCommand(versionCmd)
	tbbCmd.AddCommand(balancesCmd())
	tbbCmd.AddCommand(runCmd())
	tbbCmd.AddCommand(migrateCmd())

	err := tbbCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addDefaultRequiredFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagDataDir, "", "absolute path of data where db will be stored")
	cmd.MarkFlagRequired(flagDataDir)
}
func getDataDirFromCmd(cmd *cobra.Command) string {
	dataDir, _ := cmd.Flags().GetString(flagDataDir)

	return fs.ExpandPath(dataDir)
}
func incorrectUsageErr() error {
	return fmt.Errorf("incorrect usage")
}
