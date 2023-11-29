package blocktree

import "github.com/google/uuid"

type SpaceID = uuid.UUID

type Space struct {
	ID   SpaceID
	Name string
}
