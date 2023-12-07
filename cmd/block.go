package cmd

import (
	"context"
	"fmt"
	v1 "github.com/emrgen/blocktree/apis/v1"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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
	var spaceID, blockID, refID, pos, object string
	var insertCmd = &cobra.Command{
		Use:   "insert",
		Short: "Insert a block",
		Run: func(cmd *cobra.Command, args []string) {
			if spaceID == "" {
				panic("space ID is required")
			}
			spaceID = sanitizeID(spaceID)
			if blockID == "" {
				panic("block ID is required")
			}
			blockID = sanitizeID(blockID)

			if refID == "" {
				panic("ref ID is required")
			}
			refID = sanitizeID(refID)

			switch pos {
			case "start":
				pos = v1.PointerPosition_POINTER_POSITION_START.String()
			case "end":
				pos = v1.PointerPosition_POINTER_POSITION_END.String()
			case "before":
				pos = v1.PointerPosition_POINTER_POSITION_BEFORE.String()
			case "after":
				pos = v1.PointerPosition_POINTER_POSITION_AFTER.String()
			default:
				pos = v1.PointerPosition_POINTER_POSITION_END.String()
			}

			if object == "" {
				object = "para"
			}

			tx := &v1.Transaction{
				TransactionId: uuid.New().String(),
				SpaceId:       spaceID,
				UserId:        uuid.Nil.String(),
				Ops: []*v1.Op{{
					Table:   "block",
					BlockId: blockID,
					Type:    v1.OpType_OP_TYPE_INSERT,
					At: &v1.Pointer{
						BlockId:  refID,
						Position: v1.PointerPosition(v1.PointerPosition_value[pos]),
					},
					Object: &object,
					Linked: nil,
					Props:  nil,
					Patch:  nil,
				}},
			}

			conn, err := createConnection(":1000")
			if err != nil {
				panic(err)
				return
			}
			defer conn.Close()

			client := v1.NewBlocktreeClient(conn)

			logrus.Infof("Creating a block: %v", tx)
			res, err := client.ApplyTransactions(context.Background(), &v1.ApplyTransactionRequest{
				Transactions: []*v1.Transaction{tx},
			})
			if err != nil {
				logrus.Infof("Failed to create a block: %v", err)
				return
			}

			logrus.Infof("Created a block: %v", res)
		},
	}

	insertCmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space ID")
	insertCmd.Flags().StringVarP(&blockID, "block", "b", "", "Block ID")
	insertCmd.Flags().StringVarP(&refID, "ref", "r", "", "Ref ID")
	insertCmd.Flags().StringVarP(&pos, "pos", "p", "", "Position")
	insertCmd.Flags().StringVarP(&object, "object", "o", "", "Object")

	return insertCmd
}

func newBlockGetCmd() *cobra.Command {
	var spaceID, blockID string
	var getCmd = &cobra.Command{
		Use:   "get",
		Short: "Get a block data",
		Run: func(cmd *cobra.Command, args []string) {
			if spaceID == "" {
				panic("space ID is required")
			}
			spaceID = sanitizeID(spaceID)

			if blockID == "" {
				panic("block ID is required")
			}
			blockID = sanitizeID(blockID)

			conn, err := createConnection(":1000")
			if err != nil {
				panic(err)
			}
			defer conn.Close()

			client := v1.NewBlocktreeClient(conn)
			logrus.Infof("Getting a block: %v", blockID)
			res, err := client.GetBlock(context.Background(), &v1.GetBlockRequest{
				SpaceId: &spaceID,
				BlockId: blockID,
			})

			if err != nil {
				logrus.Infof("Failed to get a block: %v", err)
				return
			}

			logrus.Infof("Got a block: %v", res.Block)
		},
	}

	getCmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space ID")
	getCmd.Flags().StringVarP(&blockID, "block", "b", "", "Block ID")

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
