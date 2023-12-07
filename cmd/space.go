package cmd

import (
	"context"
	v1 "github.com/emrgen/blocktree/apis/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newSpaceCmd() *cobra.Command {
	var spaceCmd = &cobra.Command{
		Use:   "space",
		Short: "Manage spaces",
	}

	spaceCmd.AddCommand(newSpaceInsertCmd())

	return spaceCmd
}

func newSpaceInsertCmd() *cobra.Command {
	var spaceID, name string
	var insertCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a space",
		Run: func(cmd *cobra.Command, args []string) {
			if spaceID == "" {
				panic("space ID is required")
			}

			spaceID = sanitizeID(spaceID)

			if name == "" {
				panic("name is required")
			}

			conn, err := createConnection(":1000")
			if err != nil {
				panic(err)
			}
			defer conn.Close()

			client := v1.NewBlocktreeClient(conn)

			logrus.Infof("Creating a space: %s", spaceID)
			res, err := client.CreateSpace(context.Background(), &v1.CreateSpaceRequest{
				SpaceId: spaceID,
				Name:    name,
			})
			if err != nil {
				panic(err)
			}

			logrus.Infof("Created space: %s", res.SpaceId)
		},
	}

	insertCmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space ID")
	insertCmd.Flags().StringVarP(&name, "name", "n", "", "Space name")

	return insertCmd
}
