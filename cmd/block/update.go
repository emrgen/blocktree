package block

import (
	"fmt"
	"github.com/emrgen/blocktree/cmd/cli"

	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a block",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("update called")
	},
}

func init() {
	cli.RootCmd.AddCommand(updateCmd)
}
