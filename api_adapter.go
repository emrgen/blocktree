package blocktree

import (
	"fmt"
	v1 "github.com/emrgen/blocktree/apis/v1"
	"github.com/google/uuid"
	"time"
)

func BlockToProtoV1(b *Block) *v1.Block {
	return &v1.Block{
		Object:   b.Type,
		BlockId:  b.ID.String(),
		ParentId: b.ParentID.String(),
	}
}

func BlockViewToProtoV1(b *BlockView) *v1.Block {
	return &v1.Block{
		Object:   b.Type,
		BlockId:  b.ID.String(),
		ParentId: b.ParentID.String(),
	}
}

func TransactionFromProtoV1(txv1 *v1.Transaction) (*Transaction, error) {
	id, _ := uuid.Parse(txv1.TransactionId)
	spaceID, _ := uuid.Parse(txv1.SpaceId)
	userID, _ := uuid.Parse(txv1.UserId)
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
		Time:    time.Now(),
		Ops:     ops,
	}

	return tx, nil
}

func OpFromProtoV1(v1op *v1.Op) (*Op, error) {
	blockID, err := uuid.Parse(v1op.BlockId)
	if err != nil {
		return nil, err
	}

	atBlockID, err := uuid.Parse(v1op.At.BlockId)
	if err != nil {
		return nil, err
	}

	if v1op.Table == "" {
		return nil, fmt.Errorf("invalid op type without table")
	}

	op := &Op{
		Type:    OpType(v1op.Type.String()),
		BlockID: blockID,
		Table:   v1op.Table,
		At: &Pointer{
			BlockID:  atBlockID,
			Position: PointerPosition(v1op.At.Position.String()),
		},
	}

	if v1op.Object != nil {
		op.Object = *v1op.Object
	}

	if v1op.Linked != nil {
		op.Linked = *v1op.Linked
	}

	if v1op.Props != nil {
		op.Props = make([]OpProp, 0)
		for _, prop := range v1op.Props {
			op.Props = append(op.Props, OpProp{
				Path:  prop.Path,
				Value: prop.Value,
			})
		}
	}

	if v1op.Patch != nil {
		op.Patch = []byte(*v1op.Patch)
	}

	return op, nil
}
