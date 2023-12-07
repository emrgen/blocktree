package space

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInsertCmd() *cobra.Command {
	//var blockID, spaceID string
	var insertCmd = &cobra.Command{
		Use:   "insert",
		Short: "Insert a block",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("insert called")
		},
	}

	//insertCmd.Flags().StringVarP("name", "n", "", "Name of the block")

	return insertCmd
}
