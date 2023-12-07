package block

import (
	"github.com/spf13/cobra"
)

func NewBlockCmd() *cobra.Command {
	var blockCmd = &cobra.Command{
		Use:   "block",
		Short: "Manage blocks",
	}

	blockCmd.AddCommand(newInsertCmd())

	return blockCmd
}
