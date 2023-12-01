package blocktree

type BlockStore interface {
	// GetBlock returns a block by id
	GetBlock(id BlockID) (*Block, error)
	GetBlocks(ids []BlockID) ([]*Block, error)
}

type TransactionStore interface {
}

type JsonDocStore interface {
}

type Store interface {
	BlockStore
	TransactionStore
	JsonDocStore
}
