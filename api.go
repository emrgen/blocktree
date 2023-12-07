package blocktree

import (
	"context"
	"errors"
	v1 "github.com/emrgen/blocktree/apis/v1"
	"github.com/google/uuid"
)

var _ v1.BlocktreeServer = (*Api)(nil)

type Api struct {
	v1.BlocktreeServer
	store Store
}

func NewApi(store Store) *Api {
	return &Api{
		store: store,
	}
}

// ApplyTransactions applies a list of transactions to the blocktree store
func (a *Api) ApplyTransactions(ctx context.Context, req *v1.ApplyTransactionRequest) (*v1.ApplyTransactionResponse, error) {
	txs := req.GetTransactions()
	transactions := make([]*Transaction, 0, len(txs))
	for _, tx := range txs {
		//logrus.Info("Parsing transaction: ", tx)
		transaction, err := TransactionFromProtoV1(tx)
		if err != nil {
			return nil, err
		}
		//logrus.Info("Parsed transaction: ", transaction)
		transactions = append(transactions, transaction)
	}

	res := &v1.ApplyTransactionResponse{
		Transactions: make([]*v1.ApplyTransactionResult, len(transactions)),
	}

	for _, tx := range transactions {
		//logrus.Info("Applying transaction: ", tx)
		change, err := tx.Prepare(a.store)
		if err != nil {
			//logrus.Info("Failed to prepare transaction: ", err)
			if errors.Is(err, ErrDetectedCycle) || errors.Is(err, ErrCreatesCycle) {
				continue
			}

			return nil, err
		}
		//logrus.Info("Prepared transaction: ", change)
		err = a.store.Apply(&tx.SpaceID, change)
		//logrus.Infof("Applied transaction: %s", tx.ID)
		if err != nil {
			return nil, err
		}
	}

	//logrus.Info("Applied transactions: ", transactions)
	return res, nil
}

// CreateSpace creates a new space in the blocktree store
func (a *Api) CreateSpace(ctx context.Context, req *v1.CreateSpaceRequest) (*v1.CreateSpaceResponse, error) {
	spaceID, _ := uuid.Parse(req.GetSpaceId())
	err := a.store.CreateSpace(&Space{
		ID:   spaceID,
		Name: req.GetName(),
	})
	if err != nil {
		return nil, err
	}

	return &v1.CreateSpaceResponse{
		SpaceId: spaceID.String(),
	}, nil
}

func (a *Api) GetBlock(ctx context.Context, req *v1.GetBlockRequest) (*v1.GetBlockResponse, error) {
	blockID, _ := uuid.Parse(req.GetBlockId())
	spaceID := &uuid.Nil
	var err error
	if req.GetSpaceId() == "" {
		// get space id from block id
		spaceID, err = a.store.GetBlockSpaceID(&blockID)
		if err != nil {
			return nil, err
		}
	}

	block, err := a.store.GetBlock(spaceID, blockID)
	if err != nil {
		return nil, err
	}

	return &v1.GetBlockResponse{
		Block: BlockToProtoV1(block),
	}, nil
}

func (a *Api) GetChildrenBlocks(ctx context.Context, req *v1.GetBlockChildrenRequest) (*v1.GetBlockChildrenResponse, error) {
	blockID, _ := uuid.Parse(req.GetBlockId())
	spaceID := &uuid.Nil
	var err error
	if req.GetSpaceId() == "" {
		// get space id from block id
		spaceID, err = a.store.GetBlockSpaceID(&blockID)
		if err != nil {
			return nil, err
		}
	}

	blocks, err := a.store.GetChildrenBlocks(spaceID, blockID)
	if err != nil {
		return nil, err
	}

	v1blocks := make([]*v1.Block, len(blocks))
	for i, block := range blocks {
		v1blocks[i] = BlockToProtoV1(block)
	}

	return &v1.GetBlockChildrenResponse{
		Blocks: v1blocks,
	}, nil
}

func (a *Api) GetDescendantBlocks(ctx context.Context, req *v1.GetBlockDescendantsRequest) (*v1.GetBlockDescendantsResponse, error) {
	blockID, _ := uuid.Parse(req.GetBlockId())
	spaceID := &uuid.Nil
	var err error
	if req.GetSpaceId() == "" {
		// get space id from block id
		spaceID, err = a.store.GetBlockSpaceID(&blockID)
		if err != nil {
			return nil, err
		}
	}

	blocks, err := a.store.GetDescendantBlocks(spaceID, blockID)
	if err != nil {
		return nil, err
	}

	view, err := BlockViewFromBlocks(blockID, blocks)
	if err != nil {
		return nil, err
	}

	return &v1.GetBlockDescendantsResponse{
		Block: BlockViewToProtoV1(view),
	}, nil
}
