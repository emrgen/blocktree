package cmd

import (
	"context"

	v1 "github.com/emrgen/blocktree/apis/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func newPageCmd() *cobra.Command {
	var pageCmd = &cobra.Command{
		Use:   "page",
		Short: "Manage pages",
	}

	//pageCmd.AddCommand(newPageInsertCmd())
	pageCmd.AddCommand(newPageGetCmd())
	pageCmd.AddCommand(newPageSubPagesCmd())
	//pageCmd.AddCommand(newPageUpdateCmd())
	//pageCmd.AddCommand(newPageDeleteCmd())

	return pageCmd
}

func newPageGetCmd() *cobra.Command {
	var pageID, spaceID string
	var getCmd = &cobra.Command{
		Use:   "get",
		Short: "Get a page",
		Run: func(cmd *cobra.Command, args []string) {
			if spaceID == "" {
				panic("space ID is required")
			}
			spaceID = sanitizeID(spaceID)

			if pageID == "" {
				panic("page ID is required")
			}
			pageID = sanitizeID(pageID)

			conn, err := grpc.Dial(":4100", grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				logrus.Fatal(err)
			}
			defer conn.Close()
			client := v1.NewBlocktreeClient(conn)

			req := &v1.GetBlockDescendantsRequest{
				BlockId: pageID,
			}
			if spaceID != "" {
				req.SpaceId = &spaceID
			}

			logrus.Infof("Getting page %v", req)

			getPage, err := client.GetDescendants(context.Background(), req)
			if err != nil {
				logrus.Fatal(err)
				return
			}

			printBlock(getPage.Block)
		},
	}

	getCmd.Flags().StringVarP(&pageID, "page", "b", "", "Page ID")
	getCmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space ID")

	return getCmd
}

func newPageSubPagesCmd() *cobra.Command {
	var pageID, spaceID string
	var getCmd = &cobra.Command{
		Use:   "subpages",
		Short: "Get a page",
		Run: func(cmd *cobra.Command, args []string) {
			if spaceID == "" {
				panic("space ID is required")
			}
			spaceID = sanitizeID(spaceID)

			if pageID == "" {
				panic("page ID is required")
			}
			pageID = sanitizeID(pageID)

			conn, err := grpc.Dial(":4100", grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				logrus.Fatal(err)
			}
			defer conn.Close()
			client := v1.NewBlocktreeClient(conn)

			logrus.Info("Getting a page")
			req := &v1.GetBlockChildrenRequest{
				BlockId: pageID,
			}
			if spaceID != "" {
				req.SpaceId = &spaceID
			}

			getPage, err := client.GetChildren(context.Background(), req)
			if err != nil {
				logrus.Fatal(err)
				return
			}

			tree := treeprint.New()
			tree.AddNode(pageID)
			branch := tree.AddBranch("children")
			for _, block := range getPage.Blocks {
				branch.AddNode(block.BlockId)
			}

			logrus.Info(tree.String())
		},
	}

	getCmd.Flags().StringVarP(&pageID, "page", "b", "", "Page ID")
	getCmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space ID")

	return getCmd
}
