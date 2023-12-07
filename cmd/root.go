package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "blocktree",
	Short: "Manage a blocktree",
	// Run: func(cmd *cobra.Command, args []string) { },
}

func init() {
	rootCmd.AddCommand(newServeCmd())
	rootCmd.AddCommand(newSpaceCmd())
	rootCmd.AddCommand(newBlockCmd())
	rootCmd.AddCommand(newPageCmd())
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
