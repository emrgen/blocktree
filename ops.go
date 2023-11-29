package blocktree

import "github.com/google/uuid"

type BlockOps struct {
	Ops []Op
}

type OpType string

const (
	OpTypeCreate OpType = "insert"
	OpTypeUpdate OpType = "update" // update properties
	OpTypeMove   OpType = "move"
	OpTypeLink   OpType = "link"
	OpTypeUnlink OpType = "unlink"
	OpTypeDelete OpType = "delete"
	OpTypeErase  OpType = "erase"
)

type Position string

const (
	PositionBefore Position = "before"
	PositionAfter  Position = "after"
	PositionFirst  Position = "start"
	PositionLast   Position = "end"
	PositionInside Position = "inside"
)

type Pointer struct {
	BlockID  uuid.UUID
	Position Position
}

type OpBlock struct {
	ID         uuid.UUID
	Type       *string
	Properties map[string]interface{}
}

// Op is an operation that is applied to a blocktree.
type Op struct {
	Type  OpType  `json:"type"`
	Block OpBlock `json:"block"`
	At    Pointer `json:"at"`
}
