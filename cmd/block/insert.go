package block

import (
	"fmt"
	"github.com/spf13/cobra"
)

func newInsertCmd() *cobra.Command {
	var insertCmd = &cobra.Command{
		Use:   "insert",
		Short: "Insert a block",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("insert called")
		},
	}

	return insertCmd
}
