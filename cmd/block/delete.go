package block

import (
	"fmt"
	"github.com/emrgen/blocktree/cmd/cli"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a block",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("delete called")
	},
}

func init() {
	cli.RootCmd.AddCommand(deleteCmd)
}
