package blocktree

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/google/uuid"
)

var (
	ErrCreatesCycle    = fmt.Errorf("operation creates cycle")
	ErrDetectedCycle   = fmt.Errorf("existing cycle detected")
	ErrFailedToPublish = fmt.Errorf("failed to publish sync blocks")
)

type TransactionID = uuid.UUID

// Transaction is a collection of Ops that are applied to a Space.
type Transaction struct {
	ID      TransactionID
	SpaceID SpaceID
	UserID  uuid.UUID
	Time    time.Time
	Ops     []Op
	changes *SyncBlocks
}

// prepare prepares the transaction for application to the store.
// changes are applied to the store in one transaction.
func (tx *Transaction) prepare(store Store) (*storeChange, error) {
	//check if transaction is not already applied
	_, err := store.GetTransaction(&tx.SpaceID, tx.ID)
	// the transaction is already applied and exists in the store
	if err == nil {
		return &storeChange{
			blockChange:   newBlockChange(),
			jsonDocChange: nil,
			tx:            tx,
		}, nil
	}
	//transaction, err := store.GetLatestTransaction(&tx.SpaceID)
	//if err != nil {
	//	return nil, err
	//}
	//if transaction != nil && transaction.Version >= tx.Version {
	//	return nil, fmt.Errorf("transaction already applied")
	//}

	//check if transaction is valid (no cycles, etc)

	if tx.Ops == nil || len(tx.Ops) == 0 {
		return nil, errors.New("transaction has no ops")
	}

	// load the referenced blocks
	existingBlockIDs, _ := tx.relevantBlockIDs()
	if cycle, err := tx.createsCycles(store, existingBlockIDs); err != nil {
		return nil, err
	} else if cycle {
		return nil, fmt.Errorf("transaction creates cycles")
	}

	relevantBlocks, err := store.GetBlocks(&tx.SpaceID, existingBlockIDs.ToSlice())
	if err != nil {
		return nil, err
	}

	if len(relevantBlocks) != existingBlockIDs.Size() {
		logrus.Infof("relevant blocks: %v", existingBlockIDs.ToSlice())
		return nil, fmt.Errorf("cannot find all referenced blocks")
	}
	stage := newStageTable()
	//logrus.Infof("relevant blocks: %v", existingBlockIDs.ToSlice())
	for _, block := range relevantBlocks {
		stage.add(block)
	}

	// check and load relevant blocks from the store to the stage
	for _, op := range tx.Ops {
		switch {
		case op.Type == OpTypeInsert:
			if op.At == nil {
				return nil, fmt.Errorf("invalid create op without at: %v", op)
			}

			//check if block has type prop
			if op.Object == "" {
				return nil, fmt.Errorf("invalid create op without type: %v", op)
			}

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
					block, err := op.IntoBlock(blocks[0].ID)
					if err != nil {
						return nil, err
					}
					stage.park(block)
				}
			case op.At.Position == PositionInside:
				if op.Linked {
					blocks, err := tx.loadRelevantBlocks(store, &op)
					if err != nil {
						return nil, err
					}
					if len(blocks) < 1 {
						return nil, fmt.Errorf("cannot find referenced block for linking: %v", op)
					}
					for _, block := range blocks {
						stage.add(block)
					}
					block, err := op.IntoBlock(op.At.BlockID)
					if err != nil {
						return nil, err
					}
					stage.park(block)
				} else {
					return nil, fmt.Errorf("cannot insert inside a block: %v", op)
				}
			}
		case op.Type == OpTypeMove:
			if op.ParentID == nil {
				return nil, fmt.Errorf("invalid move op without parent id: %v", op)
			}

			if op.At == nil {
				return nil, fmt.Errorf("invalid move op without at: %v", op)
			}

			if op.At.BlockID == op.BlockID {
				return nil, fmt.Errorf("invalid move op with same block id: %v", op)
			}

			// load blocks old parent
			// if the block is inserted in this transaction, it is not in the store yet
			// its parent can also be inserted in this transaction,
			// so we need to check the stage first
			parked, ok := stage.parked(op.BlockID)
			var parent *Block
			if ok {
				if parked.ParentID == uuid.Nil {
					return nil, fmt.Errorf("newly inserted block has no parent id set: %v", op)
				}
				parkedParent, ok := stage.parked(parked.ParentID)
				if ok {
					parent = parkedParent
				} else {
					parent, err = store.GetBlock(&tx.SpaceID, parked.ParentID)
					if err != nil {
						return nil, err
					}
				}
			} else {
				parent, err = store.GetParentBlock(&tx.SpaceID, op.BlockID)
				if err != nil {
					return nil, err
				}
			}

			stage.add(parent)

			if _, ok := stage.parked(op.At.BlockID); ok {
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

				block := NewBlock(op.BlockID, parent.ID, "")
				stage.add(block)
			case op.At.Position == PositionStart || op.At.Position == PositionEnd:
				blocks, err := tx.loadRelevantBlocks(store, &op)
				if err != nil {
					return nil, err
				}

				if len(blocks) < 1 {
					return nil, fmt.Errorf("cannot find referenced block for move start/end: %v", op)
				}
				for _, block := range blocks {
					stage.add(block)
				}

				_, ok := stage.parked(op.BlockID)
				if ok {
					continue
				}

				block := NewBlock(op.BlockID, parent.ID, "")
				stage.add(block)
			case op.At.Position == PositionInside:
				return nil, fmt.Errorf("cannot move inside a block: %v", op)
			}
		case op.Type == OpTypeUpdate || op.Type == OpTypePatch || op.Type == OpTypeDelete || op.Type == OpTypeErase || op.Type == OpTypeUndelete || op.Type == OpTypeRestore:
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
		case op.Type == OpTypeLink:
			parent, err := store.GetBlock(&tx.SpaceID, op.At.BlockID)
			if err != nil {
				return nil, err
			}
			block, err := store.GetBlock(&tx.SpaceID, op.BlockID)
			if err != nil {
				return nil, err
			}
			stage.add(parent)
			stage.add(block)
		case op.Type == OpTypeUnlink:
			block, err := store.GetBlock(&tx.SpaceID, op.BlockID)
			if err != nil {
				return nil, err
			}
			stage.add(block)
		}
	}

	logrus.Debugf("applying transaction %v", tx.ID)
	change, err := stage.Apply(tx)
	if err != nil {
		return nil, err
	}

	return &storeChange{
		blockChange:   change,
		jsonDocChange: nil,
		tx:            tx,
	}, nil
}

// relevantBlocks returns a set of preexisting block ids that are referenced by the transaction
func (tx *Transaction) relevantBlockIDs() (*Set[BlockID], *Set[BlockID]) {
	relevant := NewSet[uuid.UUID]()
	inserted := NewSet[uuid.UUID]()

	for _, op := range tx.Ops {
		logrus.Debugf("op: %v", op)
		if op.Type == OpTypeInsert {
			if op.At != nil {
				if !inserted.Contains(op.At.BlockID) {
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

	return relevant, inserted
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

	moveTree := newMoveTree(tx.SpaceID)
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
				err := moveTree.move(op.BlockID, *parentID)
				if err != nil {
					if errors.Is(ErrDetectedCycle, err) {
						return true, nil
					}
					return false, err
				}
			case op.At.Position == PositionStart || op.At.Position == PositionEnd:
				err := moveTree.move(op.BlockID, op.At.BlockID)
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
	relevantBlocks := make([]*Block, 0)
	// load the referenced blocks
	switch {
	case op.Type == OpTypeInsert || op.Type == OpTypeMove:
		switch {
		case op.At.Position == PositionAfter:
			blocks, err := store.GetParentWithNextBlock(&tx.SpaceID, op.At.BlockID)
			//logrus.Info("get parent with next block ", blocks)
			if err != nil {
				logrus.Infof("get parent with next block error: %v", err)
				return nil, err
			}
			relevantBlocks = append(relevantBlocks, blocks...)
		case op.At.Position == PositionBefore:
			blocks, err := store.GetParentWithPrevBlock(&tx.SpaceID, op.At.BlockID)
			if err != nil {
				logrus.Infof("get parent with prev block error: %v", err)
				return nil, err
			}
			relevantBlocks = append(relevantBlocks, blocks...)
		case op.At.Position == PositionStart:
			blocks, err := store.GetWithFirstChildBlock(&tx.SpaceID, op.At.BlockID)
			if err != nil {
				return nil, err
			}
			relevantBlocks = append(relevantBlocks, blocks...)
		case op.At.Position == PositionEnd:
			blocks, err := store.GetWithLastChildBlock(&tx.SpaceID, op.At.BlockID)
			if err != nil {
				return nil, err
			}
			relevantBlocks = append(relevantBlocks, blocks...)
		case op.At.Position == PositionInside:
			block, err := store.GetParentBlock(&tx.SpaceID, op.At.BlockID)
			if err != nil {
				return nil, err
			}
			relevantBlocks = append(relevantBlocks, block)
		}
	case op.Type == OpTypeLink:
		block, err := store.GetBlock(&tx.SpaceID, op.At.BlockID)
		if err != nil {
			return nil, err
		}
		relevantBlocks = append(relevantBlocks, block)
	}

	return relevantBlocks, nil
}

// OpType is the type of operation
type OpType string

const (
	OpTypeInsert   OpType = "insert"
	OpTypeMove     OpType = "move"
	OpTypeUpdate   OpType = "update" // update properties of a block
	OpTypePatch    OpType = "patch"  // patch json document of a block
	OpTypeLink     OpType = "link"
	OpTypeUnlink   OpType = "unlink"
	OpTypeDelete   OpType = "delete"
	OpTypeUndelete OpType = "undelete"
	OpTypeErase    OpType = "erase"
	OpTypeRestore  OpType = "restore"
)

type PointerPosition string

const (
	// PositionBefore before and after are used for inserting blocks before or after a reference block
	PositionBefore PointerPosition = "before"
	PositionAfter  PointerPosition = "after"
	// PositionStart start and end are used for inserting blocks at the start or end of a reference block children
	PositionStart PointerPosition = "start"
	PositionEnd   PointerPosition = "end"
	// PositionInside inside is used for inserting blocks inside a reference block
	PositionInside PointerPosition = "inside"
)

// Pointer is a position wrt a block
type Pointer struct {
	BlockID  uuid.UUID       `json:"block_id"`
	Position PointerPosition `json:"position"`
}

// OpProp is a property operation
type OpProp struct {
	Path  []string
	Value interface{}
}

// Op is an operation that is applied to a blocktree.
type Op struct {
	Table    string   `json:"table"`
	Type     OpType   `json:"type"`
	Object   string   `json:"object"`
	Linked   bool     `json:"linked"`
	BlockID  BlockID  `json:"block_id"`
	ParentID *BlockID `json:"parent_id"` // parent_id before move
	At       *Pointer `json:"at"`
	Props    []byte   `json:"props"`
	Patch    []byte   `json:"patch"`
}

// IntoBlock converts the operation into a block object
func (op *Op) IntoBlock(parentID ParentID) (*Block, error) {
	if op.Type != OpTypeInsert {
		return nil, fmt.Errorf("op is not a insert op")
	}

	// insert op must have a block object
	if op.Object == "" {
		return nil, fmt.Errorf("insert op is missing block object type")
	}

	if op.Table == "" {
		return nil, fmt.Errorf("insert op is missing table")
	}

	jsonDoc := DefaultJsonDoc()
	if op.Patch != nil {
		err := jsonDoc.Apply(op.Patch)
		if err != nil {
			return nil, err
		}
	}

	return &Block{
		ParentID: parentID,
		Type:     op.Object,
		Table:    op.Table,
		ID:       op.BlockID,
		Index:    DefaultFracIndex(),
		Json:     jsonDoc,
		Deleted:  false,
		Erased:   false,
		Linked:   op.Linked,
	}, nil
}

// IntoProp converts the operation into a property operation
func (op *Op) String() string {
	switch op.Type {
	case OpTypeInsert:
		return fmt.Sprintf("%s %s %s %s %s", op.Type, op.BlockID, op.At.Position, op.At.BlockID, op.Props)
	}
	return fmt.Sprintf("%s %s %s", op.Type, op.BlockID, op.At)
}
