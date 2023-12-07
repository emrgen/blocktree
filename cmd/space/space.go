package space

import (
	"github.com/spf13/cobra"
)

func NewSpaceCmd() *cobra.Command {
	var spaceCmd = &cobra.Command{
		Use:   "space",
		Short: "Manage spaces",
	}

	spaceCmd.AddCommand(newInsertCmd())

	return spaceCmd
}
