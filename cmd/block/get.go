package block

import (
	"fmt"
	"github.com/emrgen/blocktree/cmd/cli"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a block data",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("get called")
	},
}

func init() {
	cli.RootCmd.AddCommand(getCmd)
}
