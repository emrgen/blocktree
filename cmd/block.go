package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func newBlockCmd() *cobra.Command {
	var blockCmd = &cobra.Command{
		Use:   "block",
		Short: "Manage blocks",
	}

	blockCmd.AddCommand(newBlockInsertCmd())
	blockCmd.AddCommand(newBlockGetCmd())
	blockCmd.AddCommand(newBlockUpdateCmd())
	blockCmd.AddCommand(newBlockDeleteCmd())

	return blockCmd
}

func newBlockInsertCmd() *cobra.Command {
	var insertCmd = &cobra.Command{
		Use:   "insert",
		Short: "Insert a block",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("insert called")
		},
	}

	return insertCmd
}

func newBlockGetCmd() *cobra.Command {
	var getCmd = &cobra.Command{
		Use:   "get",
		Short: "Get a block data",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("get called")
		},
	}

	return getCmd
}

func newBlockUpdateCmd() *cobra.Command {
	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update a block",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("update called")
		},
	}

	return updateCmd
}

func newBlockDeleteCmd() *cobra.Command {
	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a block",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("delete called")
		},
	}

	return deleteCmd
}
