package blocktree

import (
	"time"

	"github.com/google/uuid"
)

type TxID uuid.UUID

// Transaction is a collection of Ops that are applied to a Space.
type Transaction struct {
	ID        uuid.UUID
	SpaceID   uuid.UUID
	UserID    uuid.UUID
	Time      time.Time
	TxCounter int64
	Ops       []Op
}
