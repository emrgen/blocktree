package cli

import (
	"github.com/emrgen/blocktree/cmd/block"
	"github.com/emrgen/blocktree/cmd/server"
	"github.com/emrgen/blocktree/cmd/space"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "blocktree",
	Short: "Manage a blocktree",
	// Run: func(cmd *cobra.Command, args []string) { },
}

func init() {
	rootCmd.AddCommand(server.NewServeCmd())
	rootCmd.AddCommand(space.NewSpaceCmd())
	rootCmd.AddCommand(block.NewBlockCmd())

}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
