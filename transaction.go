package blocktree

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCreatesCycle  = fmt.Errorf("cycle detected")
	ErrDetectedCycle = fmt.Errorf("cycle detected")
)

type TransactionID = uuid.UUID

// Transaction is a collection of Ops that are applied to a Space.
type Transaction struct {
	ID        TransactionID `json:"id"`
	SpaceID   SpaceID
	UserID    uuid.UUID
	Time      time.Time
	TxCounter int64
	Ops       []Op
}

func (tx *Transaction) Prepare(store Store) (*StoreChange, error) {
	// check if transaction is not already applied
	transaction, err := store.GetTransaction(&tx.SpaceID, &tx.ID)
	if err != nil {
		return nil, err
	}
	if transaction != nil {
		return nil, fmt.Errorf("transaction already applied")
	}

	//check if transaction is valid (no cycles, etc)

	// load the referenced blocks
	relevantBlockIDs := tx.relevantBlockIDs()
	if cycle, err := tx.createsCycles(store, relevantBlockIDs); err != nil {
		return nil, err
	} else if cycle {
		return nil, fmt.Errorf("transaction creates cycles")
	}

	relevantBlocks, err := store.GetBlocks(&tx.SpaceID, relevantBlockIDs.ToSlice())
	if err != nil {
		return nil, err
	}

	if len(relevantBlocks) != relevantBlockIDs.Cardinality() {
		return nil, fmt.Errorf("cannot find all referenced blocks")
	}
	stage := NewStageTable()
	for _, block := range relevantBlocks {
		stage.add(block)
	}

	// load all relevant blocks into the stage
	for _, op := range tx.Ops {
		switch {
		case op.Type == OpTypeInsert:
			if op.At == nil {
				return nil, fmt.Errorf("invalid create op without at: %v", op)
			}

			//check if block has type prop
			if typ, ok := op.Props["type"]; !ok {
				return nil, fmt.Errorf("invalid create op without type: %v", op)
			} else if _, ok := typ.(string); !ok {
				return nil, fmt.Errorf("invalid create op with non-string type: %v", op)
			}

			switch {
			case op.At.Position == PositionAfter || op.At.Position == PositionBefore:
				refBlock, ok := stage.parked(op.At.BlockID)
				if ok {
					block, err := op.IntoBlock(*refBlock.ParentID)
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
					block, err := op.IntoBlock(blocks[0].ID)
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
				block := NewBlock(op.BlockID, &parent.ID, "")
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

	logrus.Info("apply stage")
	change, err := stage.Apply(tx)
	if err != nil {
		return nil, err
	}

	return &StoreChange{
		blockChange:   change,
		jsonDocChange: nil,
		txChange:      nil,
	}, nil
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

// createsCycles returns true if the transaction creates cycles in the blocktree
func (tx *Transaction) createsCycles(store Store, blockIDs *Set[BlockID]) (bool, error) {
	if !tx.moves() {
		return false, nil
	}

	// load all relevant blocks
	blockEdges, err := store.GetAncestorEdges(&tx.SpaceID, blockIDs.ToSlice())
	if err != nil {
		return false, err
	}

	moveTree := NewMoveTree(tx.SpaceID)
	for _, edge := range blockEdges {
		moveTree.addEdge(edge.childID, edge.parentID)
	}

	for _, op := range tx.Ops {
		switch {
		case op.Type == OpTypeInsert:
			if op.At == nil {
				return false, fmt.Errorf("invalid create op without at: %v", op)
			}
			switch {
			case op.At.Position == PositionAfter || op.At.Position == PositionBefore:
				parentID, ok := moveTree.getParent(op.At.BlockID)
				if !ok {
					return false, fmt.Errorf("cannot find parent for insert after/before: %v", op)
				}
				moveTree.addEdge(op.BlockID, *parentID)
			case op.At.Position == PositionStart || op.At.Position == PositionEnd:
				if !moveTree.contains(op.At.BlockID) {
					return true, nil
				}
				moveTree.addEdge(op.BlockID, op.At.BlockID)
			case op.At.Position == PositionInside:
				return false, fmt.Errorf("cannot insert inside a block: %v", op)
			}
		case op.Type == OpTypeMove:
			if op.At == nil {
				return false, fmt.Errorf("invalid move op without at: %v", op)
			}
			if op.At.BlockID == op.BlockID {
				return false, fmt.Errorf("invalid move op with same block id: %v", op)
			}

			switch {
			case op.At.Position == PositionAfter || op.At.Position == PositionBefore:
				parentID, ok := moveTree.getParent(op.At.BlockID)
				if !ok {
					return false, fmt.Errorf("cannot find parent for move after/before: %v", op)
				}
				err := moveTree.Move(op.BlockID, *parentID)
				if err != nil {
					if errors.Is(ErrDetectedCycle, err) {
						return true, nil
					}
					return false, err
				}
			case op.At.Position == PositionStart || op.At.Position == PositionEnd:
				err := moveTree.Move(op.BlockID, op.At.BlockID)
				if err != nil {
					if errors.Is(ErrDetectedCycle, err) {
						return true, nil
					}
					return false, err
				}
			case op.At.Position == PositionInside:
				continue
			}
		}
	}

	return false, nil
}

func (tx *Transaction) loadRelevantBlocks(store Store, op *Op) ([]*Block, error) {
	// load the referenced blocks
	switch {
	case op.Type == OpTypeInsert || op.Type == OpTypeMove:
		switch {
		case op.At.Position == PositionAfter:
			blocks, err := store.GetParentWithNextBlock(&tx.SpaceID, op.At.BlockID)
			if err != nil {
				return nil, err
			}
			return blocks, nil
		case op.At.Position == PositionBefore:
			blocks, err := store.GetParentWithPrevBlock(&tx.SpaceID, op.At.BlockID)
			if err != nil {
				return nil, err
			}
			return blocks, nil

		case op.At.Position == PositionStart:
			blocks, err := store.GetWithFirstChildBlock(&tx.SpaceID, op.At.BlockID)
			if err != nil {
				return nil, err
			}
			return blocks, nil
		case op.At.Position == PositionEnd:
			blocks, err := store.GetWithLastChildBlock(&tx.SpaceID, op.At.BlockID)
			if err != nil {
				return nil, err
			}
			return blocks, nil
		}
	}

	return []*Block{}, nil
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
	BlockID BlockID                `json:"block_id"`
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
		ParentID: &parentID,
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
