package cmd

import (
	"context"
	"encoding/json"
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
	blockCmd.AddCommand(newBlockMoveCmd())

	blockCmd.AddCommand(newBlockGetCmd())
	blockCmd.AddCommand(newBlockUpdateCmd())
	blockCmd.AddCommand(newBlockDeleteCmd())
	blockCmd.AddCommand(newBlockUndeleteCmd())
	blockCmd.AddCommand(newBlockPatchCmd())

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

			logrus.Infof("Inserting block: %v", blockID)
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

func newBlockMoveCmd() *cobra.Command {
	var spaceID, blockID, refID, pos, object string
	var insertCmd = &cobra.Command{
		Use:   "move",
		Short: "Move a block",
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
					Type:    v1.OpType_OP_TYPE_MOVE,
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

			logrus.Infof("Moving block: %v", blockID)
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
				logrus.Infof("space ID is required")
				return
			}
			spaceID = sanitizeID(spaceID)

			if blockID == "" {
				logrus.Infof("block ID is required")
				return
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

			msg, _ := json.Marshal(res.Block)
			logrus.Infof("Got a block: %v", string(msg))
		},
	}

	getCmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space ID")
	getCmd.Flags().StringVarP(&blockID, "block", "b", "", "Block ID")

	return getCmd
}

func newBlockUpdateCmd() *cobra.Command {
	var spaceID, blockID, patch string
	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update a block",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("update called")
		},
	}

	updateCmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space ID")
	updateCmd.Flags().StringVarP(&blockID, "block", "b", "", "Block ID")
	updateCmd.Flags().StringVarP(&patch, "patch", "p", "", "Patch")

	return updateCmd
}

func newBlockDeleteCmd() *cobra.Command {
	var spaceID, blockID string
	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Undelete a block",
		Run: func(cmd *cobra.Command, args []string) {
			if spaceID == "" {
				logrus.Infof("space ID is required")
				return
			}
			spaceID = sanitizeID(spaceID)

			if blockID == "" {
				logrus.Infof("block ID is required")
				return
			}
			blockID = sanitizeID(blockID)

			conn, err := createConnection(":1000")
			if err != nil {
				panic(err)
			}
			defer conn.Close()

			client := v1.NewBlocktreeClient(conn)
			logrus.Infof("Deleting block: %v", blockID)
			tx := &v1.Transaction{
				TransactionId: uuid.New().String(),
				SpaceId:       spaceID,
				UserId:        uuid.Nil.String(),
				Ops: []*v1.Op{{
					Table:   "block",
					BlockId: blockID,
					Type:    v1.OpType_OP_TYPE_UNDELETE,
				}},
			}
			res, err := client.ApplyTransactions(context.Background(), &v1.ApplyTransactionRequest{
				Transactions: []*v1.Transaction{tx},
			})

			if err != nil {
				logrus.Infof("Failed to get a block: %v", err)
				return
			}

			msg, _ := json.Marshal(res.Transactions)
			logrus.Infof("Got a block: %v", string(msg))
		},
	}

	deleteCmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space ID")
	deleteCmd.Flags().StringVarP(&blockID, "block", "b", "", "Block ID")

	return deleteCmd
}

func newBlockUndeleteCmd() *cobra.Command {
	var spaceID, blockID string
	var undeleteCmd = &cobra.Command{
		Use:   "undelete",
		Short: "Undelete a block",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("undelete called")
		},
	}

	undeleteCmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space ID")
	undeleteCmd.Flags().StringVarP(&blockID, "block", "b", "", "Block ID")

	return undeleteCmd
}

func newBlockPatchCmd() *cobra.Command {
	var spaceID, blockID, patch string
	var patchCmd = &cobra.Command{
		Use:   "patch",
		Short: "Patch a block",
		Run: func(cmd *cobra.Command, args []string) {
			if spaceID == "" {
				logrus.Infof("space ID is required")
				return
			}
			spaceID = sanitizeID(spaceID)

			if blockID == "" {
				logrus.Infof("block ID is required")
				return
			}
			blockID = sanitizeID(blockID)

			if patch == "" {
				logrus.Infof("patch is required")
				return
			}

			conn, err := createConnection(":1000")
			if err != nil {
				panic(err)
			}
			defer conn.Close()

			client := v1.NewBlocktreeClient(conn)
			logrus.Infof("Patching a block: %v", blockID)
			tx := v1.Transaction{
				TransactionId: uuid.New().String(),
				SpaceId:       spaceID,
				UserId:        uuid.Nil.String(),
				Ops: []*v1.Op{{
					Table:   "block",
					BlockId: blockID,
					Type:    v1.OpType_OP_TYPE_PATCH,
					Patch:   &patch,
				}},
			}
			res, err := client.ApplyTransactions(context.Background(), &v1.ApplyTransactionRequest{
				Transactions: []*v1.Transaction{&tx},
			})

			if err != nil {
				logrus.Infof("Failed to patch a block: %v", err)
				return
			}

			logrus.Infof("Patched a block: %v", res.Transactions)
		},
	}

	patchCmd.Flags().StringVarP(&spaceID, "space", "s", "", "Space ID")
	patchCmd.Flags().StringVarP(&blockID, "block", "b", "", "Block ID")
	patchCmd.Flags().StringVarP(&patch, "patch", "p", "", "Patch")

	return patchCmd
}
