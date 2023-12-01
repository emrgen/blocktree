package blocktree

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TxID uuid.UUID

// Transaction is a collection of Ops that are applied to a Space.
type Transaction struct {
	ID        uuid.UUID `json:"id"`
	SpaceID   uuid.UUID
	UserID    uuid.UUID
	Time      time.Time
	TxCounter int64
	Ops       []Op
}

func (tx *Transaction) Apply(store Store) (*BlockChange, error) {
	// check if transaction is not already applied

	//check if transaction is valid (no cycles, etc)

	// load the referenced blocks
	blocks := make(map[uuid.UUID]*Block)
	relevantBlockIDs := tx.relevantBlockIDs()
	relevantBlocks, err := store.GetBlocks(relevantBlockIDs.ToSlice())
	if err != nil {
		return nil, err
	}
	stage := NewStageTable()
	for _, block := range relevantBlocks {
		blocks[block.ID] = block
		stage.add(block)
	}

	// load all relevant blocks into the stage
	for _, op := range tx.Ops {
		switch {
		case op.Type == OpTypeInsert:
			if op.At == nil {
				return nil, fmt.Errorf("invalid create op without at: %v", op)
			}

			// ref block is parked
			switch {
			case op.At.Position == PositionAfter || op.At.Position == PositionBefore:
				refBlock, ok := stage.parked(op.At.BlockID)
				if ok {
					block, err := op.IntoBlock(refBlock.ParentID)
					if err != nil {
						return nil, err
					}
					stage.park(block)
				} else {
					blocks, err := tx.loadRelevantBlocks(store, &op)
					if err != nil {
						return nil, err
					}
					if len(blocks) < 2 {
						return nil, fmt.Errorf("cannot find referenced block for insert after/before: %v", op)
					}
					for _, block := range blocks {
						stage.add(block)
					}
					parent := blocks[0]
					block, err := op.IntoBlock(parent.ID)
					if err != nil {
						return nil, err
					}
					stage.park(block)
				}
			case op.At.Position == PositionStart || op.At.Position == PositionEnd:
				_, ok := stage.parked(op.At.BlockID)
				if ok {
					block, err := op.IntoBlock(op.At.BlockID)
					if err != nil {
						return nil, err
					}
					stage.park(block)
				} else {
					blocks, err := tx.loadRelevantBlocks(store, &op)
					if err != nil {
						return nil, err
					}
					if len(blocks) < 1 {
						return nil, fmt.Errorf("cannot find referenced block for insert start/end: %v", op)
					}
					for _, block := range blocks {
						stage.add(block)
					}
					block, err := op.IntoBlock(blocks[0].ParentID)
					if err != nil {
						return nil, err
					}
					stage.park(block)
				}
			case op.At.Position == PositionInside:
				return nil, fmt.Errorf("cannot insert inside a block: %v", op)
			}
		case op.Type == OpTypeMove:
			if op.At == nil {
				return nil, fmt.Errorf("invalid move op without at: %v", op)
			}
			if op.At.BlockID == op.BlockID {
				return nil, fmt.Errorf("invalid move op with same block id: %v", op)
			}
			if _, ok := stage.parked(op.At.BlockID); !ok {
				continue
			}

			switch {
			case op.At.Position == PositionAfter || op.At.Position == PositionBefore:
				blocks, err := tx.loadRelevantBlocks(store, &op)
				if err != nil {
					return nil, err
				}

				if len(blocks) < 2 {
					return nil, fmt.Errorf("cannot find referenced block for move after/before: %v", op)
				}
				for _, block := range blocks {
					stage.add(block)
				}

				_, ok := stage.parked(op.BlockID)
				if ok {
					continue
				}

				parent := blocks[0]
				block := NewBlock(parent.ID, op.BlockID, "")
				stage.add(block)
			}
		case op.Type == OpTypeUpdate:
			if ok := stage.contains(op.BlockID); ok {
				continue
			}
			blocks, err := tx.loadRelevantBlocks(store, &op)
			if err != nil {
				return nil, err
			}
			if len(blocks) < 1 {
				return nil, fmt.Errorf("cannot find referenced block for update: %v", op)
			}
			for _, block := range blocks {
				stage.add(block)
			}
		case op.Type == OpTypePatch:
			panic("not implemented")
		case op.Type == OpTypeLink:
			panic("not implemented")
		case op.Type == OpTypeUnlink:
			panic("not implemented")
		case op.Type == OpTypeDelete || op.Type == OpTypeErase:
			if ok := stage.contains(op.BlockID); ok {
				continue
			}
			blocks, err := tx.loadRelevantBlocks(store, &op)
			if err != nil {
				return nil, err
			}
			if len(blocks) < 1 {
				return nil, fmt.Errorf("cannot find referenced block for delete: %v", op)
			}
			for _, block := range blocks {
				stage.add(block)
			}
		}
	}

	change, err := stage.Apply(tx)
	if err != nil {
		return nil, err
	}

	return change, nil
}

// relevantBlocks returns a set of preexisting block ids that are referenced by the transaction
func (tx *Transaction) relevantBlockIDs() *Set[BlockID] {
	relevant := NewSet[uuid.UUID]()
	inserted := NewSet[uuid.UUID]()

	for _, op := range tx.Ops {
		if op.Type == OpTypeInsert {
			if op.At != nil {
				if !relevant.Contains(op.At.BlockID) || inserted.Contains(op.At.BlockID) {
					relevant.Add(op.At.BlockID)
				}
			}
			relevant.Add(op.BlockID)
			inserted.Add(op.BlockID)
		} else {
			if op.At != nil {
				if !relevant.Contains(op.At.BlockID) || inserted.Contains(op.At.BlockID) {
					relevant.Add(op.At.BlockID)
				}
			}

			if !relevant.Contains(op.BlockID) && !inserted.Contains(op.BlockID) {
				relevant.Add(op.BlockID)
			}
		}
	}

	return relevant
}

func (tx *Transaction) moves() bool {
	for _, op := range tx.Ops {
		if op.Type == OpTypeMove {
			return true
		}
	}

	return false
}

func (tx *Transaction) createsCycles() bool {
	return false
}

func (tx *Transaction) loadRelevantBlocks(store Store, op *Op) ([]*Block, error) {
	// load the referenced blocks
	return nil, nil
}

type BlockOps struct {
	Ops []Op
}

type OpType string

const (
	OpTypeInsert OpType = "insert"
	OpTypeMove   OpType = "move"
	OpTypeUpdate OpType = "update" // update properties
	OpTypePatch  OpType = "patch"  // patch json
	OpTypeLink   OpType = "link"
	OpTypeUnlink OpType = "unlink"
	OpTypeDelete OpType = "delete"
	OpTypeErase  OpType = "erase"
)

type PointerPosition string

const (
	PositionBefore PointerPosition = "before"
	PositionAfter  PointerPosition = "after"
	PositionStart  PointerPosition = "start"
	PositionEnd    PointerPosition = "end"
	PositionInside PointerPosition = "inside"
)

// Pointer is a reference to a block position
type Pointer struct {
	BlockID  uuid.UUID       `json:"block_id"`
	Position PointerPosition `json:"position"`
}

// Op is an operation that is applied to a blocktree.
type Op struct {
	Table   string                 `json:"table"`
	Type    OpType                 `json:"type"`
	BlockID uuid.UUID              `json:"block"`
	At      *Pointer               `json:"at"`
	Props   map[string]interface{} `json:"props"`
	Patch   *JsonDocPatch          `json:"patch"`
}

func (op *Op) IntoBlock(parentID ParentID) (*Block, error) {
	if op.Type != OpTypeInsert {
		return nil, fmt.Errorf("op is not a insert op")
	}

	// op must have a type
	blockType, ok := op.Props["type"].(string)
	if !ok {
		return nil, fmt.Errorf("insert op is missing block type")
	}

	return &Block{
		ParentID: parentID,
		Type:     blockType,
		ID:       op.BlockID,
		Index:    DefaultFracIndex(),
		Props:    op.Props,
		Deleted:  false,
		Erased:   false,
	}, nil
}

func (op *Op) String() string {
	return fmt.Sprintf("%s %s %s", op.Type, op.BlockID, op.At)
}
