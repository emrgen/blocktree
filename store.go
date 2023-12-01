package blocktree

type BlockStore interface {
	// GetBlock returns a block by id
	GetBlock(id BlockID) (*Block, error)
	GetBlocks(ids []BlockID) ([]*Block, error)
	GetAncestorEdges(id []BlockID) ([]BlockEdge, error)
}

type TransactionStore interface {
	GetTransaction(id TransactionID) (*Transaction, error)
}

type JsonDocStore interface {
}

type Store interface {
	BlockStore
	TransactionStore
	JsonDocStore
}
