package blocktree

import (
	"fmt"
	"time"

	v1 "github.com/emrgen/blocktree/apis/v1"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func BlockToProtoV1(b *Block) *v1.Block {
	block := &v1.Block{
		Object:   b.Type,
		BlockId:  b.ID.String(),
		ParentId: b.ParentID.String(),
	}

	if b.Json != nil {
		content := b.Json.String()
		block.Json = &content
	}

	if b.Props != nil {
		content := b.Props.String()
		block.Props = &content
	}

	if b.Deleted {
		deleted := b.Deleted
		logrus.Infof("deleted: %t", deleted)
		block.Deleted = &deleted
	}

	if b.Erased {
		erased := b.Erased
		block.Erased = &erased
	}

	return block
}

func BlockViewToProtoV1(b *BlockView) *v1.Block {
	children := make([]*v1.Block, 0)
	links := make([]*v1.Block, 0)
	if b.Children != nil {
		for _, child := range b.Children {
			children = append(children, BlockViewToProtoV1(child))
		}
	}

	if b.Linked != nil {
		for _, link := range b.Linked {
			links = append(links, BlockViewToProtoV1(link))
		}
	}

	block := &v1.Block{
		Object:   b.Type,
		BlockId:  b.ID.String(),
		ParentId: b.ParentID.String(),
		Children: children,
		Linked:   links,
	}

	if b.Json != nil {
		content := b.Json.String()
		block.Json = &content
	}

	if b.Props != nil {
		content := b.Props.String()
		block.Props = &content
	}

	return block
}

func transactionFromProtoV1(txv1 *v1.Transaction) (*Transaction, error) {
	id, err := uuid.Parse(txv1.TransactionId)
	if err != nil {
		return nil, err
	}
	spaceID, err := uuid.Parse(txv1.SpaceId)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(txv1.UserId)
	if err != nil {
		return nil, err
	}
	ops := make([]Op, 0)
	for _, op := range txv1.Ops {
		op, err := OpFromProtoV1(op)
		if err != nil {
			return nil, err
		}
		ops = append(ops, *op)
	}

	tx := &Transaction{
		ID:      id,
		SpaceID: spaceID,
		UserID:  userID,
		Time:    time.Now(), // always use current time
		Ops:     ops,
	}

	return tx, nil
}

func OpFromProtoV1(v1op *v1.Op) (*Op, error) {
	blockID, err := uuid.Parse(v1op.BlockId)
	if err != nil {
		return nil, err
	}

	if v1op.Table == "" {
		return nil, fmt.Errorf("invalid op type without table")
	}

	opType := ""
	switch v1op.Type {
	case v1.OpType_OP_TYPE_INSERT:
		opType = "insert"
	case v1.OpType_OP_TYPE_MOVE:
		opType = "move"
	case v1.OpType_OP_TYPE_UPDATE:
		opType = "update"
	case v1.OpType_OP_TYPE_DELETE:
		opType = "delete"
	case v1.OpType_OP_TYPE_ERASE:
		opType = "erase"
	case v1.OpType_OP_TYPE_PATCH:
		opType = "patch"
	}

	if opType == "" {
		return nil, fmt.Errorf("invalid op type: %s", v1op.Type.String())
	}

	op := &Op{
		Type:    OpType(opType),
		BlockID: blockID,
		Table:   v1op.Table,
	}

	// at is required for move and insert ops
	if v1op.At == nil && (op.Type == "insert" || op.Type == "move") {
		return nil, fmt.Errorf("invalid op type %s with at", op.Type)
	}

	if v1op.At != nil {
		atBlockID, err := uuid.Parse(v1op.At.BlockId)
		if err != nil {
			return nil, err
		}

		pos := ""
		switch v1op.At.Position {
		case v1.PointerPosition_POINTER_POSITION_START:
			pos = "start"
		case v1.PointerPosition_POINTER_POSITION_END:
			pos = "end"
		case v1.PointerPosition_POINTER_POSITION_BEFORE:
			pos = "before"
		case v1.PointerPosition_POINTER_POSITION_AFTER:
			pos = "after"
		}

		if pos == "" {
			return nil, fmt.Errorf("invalid pointer position: %s", v1op.At.Position.String())
		}

		op.At = &Pointer{
			BlockID:  atBlockID,
			Position: PointerPosition(pos),
		}
	}

	if v1op.Object != nil {
		op.Object = *v1op.Object
	}

	if v1op.Linked != nil {
		op.Linked = *v1op.Linked
	}

	if v1op.Props != nil {
		op.Props = []byte(*v1op.Props)
	}

	if v1op.Patch != nil {
		op.Patch = []byte(*v1op.Patch)
	}

	return op, nil
}
