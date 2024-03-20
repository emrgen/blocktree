package blocktree

import (
	"context"

	v1 "github.com/emrgen/blocktree/apis/v1"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var _ v1.BlocktreeServer = (*grpcApi)(nil)

type grpcApi struct {
	v1.BlocktreeServer
	api *Api
}

func newGrpcApi(api *Api) *grpcApi {
	return &grpcApi{
		api: api,
	}
}

// ApplyTransactions applies a list of transactions to the blocktree store
func (a *grpcApi) ApplyTransactions(ctx context.Context, req *v1.ApplyTransactionRequest) (*v1.ApplyTransactionResponse, error) {
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

	err := a.api.Apply(transactions...)
	if err != nil {
		return nil, err
	}

	//logrus.Info("Applied transactions: ", transactions)
	return res, nil
}

// CreateSpace creates a new space in the blocktree store
func (a *grpcApi) CreateSpace(ctx context.Context, req *v1.CreateSpaceRequest) (*v1.CreateSpaceResponse, error) {
	spaceID, _ := uuid.Parse(req.GetSpaceId())
	err := a.api.CreateSpace(spaceID, req.GetName())
	if err != nil {
		return nil, err
	}

	return &v1.CreateSpaceResponse{
		SpaceId: spaceID.String(),
	}, nil
}

func (a *grpcApi) GetBlock(ctx context.Context, req *v1.GetBlockRequest) (*v1.GetBlockResponse, error) {
	blockID, _ := uuid.Parse(req.GetBlockId())
	var spaceID *uuid.UUID
	var err error
	if req.GetSpaceId() == "" {
		// get space id from block id
		spaceID, err = a.api.GetBlockSpaceID(blockID)
		if err != nil {
			return nil, err
		}
	} else {
		sid, err := uuid.Parse(req.GetSpaceId())
		if err != nil {
			return nil, err
		}
		spaceID = &sid
	}

	block, err := a.api.GetBlock(*spaceID, blockID)
	if err != nil {
		return nil, err
	}

	logrus.Infof("block %v", block)

	return &v1.GetBlockResponse{
		Block: BlockToProtoV1(block),
	}, nil
}

func (a *grpcApi) GetBlockChildren(ctx context.Context, req *v1.GetBlockChildrenRequest) (*v1.GetBlockChildrenResponse, error) {
	var err error
	var spaceID *uuid.UUID

	blockID, err := uuid.Parse(req.GetBlockId())
	if err != nil {
		return nil, err
	}

	if req.GetSpaceId() == "" {
		// get space id from block id
		spaceID, err = a.api.GetBlockSpaceID(blockID)
		if err != nil {
			return nil, err
		}
	} else {
		sid, err := uuid.Parse(req.GetSpaceId())
		if err != nil {
			return nil, err
		}
		spaceID = &sid
	}

	blocks, err := a.api.GetChildrenBlocks(*spaceID, blockID)
	if err != nil {
		return nil, err
	}

	v1blocks := make([]*v1.Block, len(blocks))
	for i, block := range blocks {
		logrus.Infof("block %v index: %s", block.ID.String(), block.Index.String())
		v1blocks[i] = BlockToProtoV1(block)
	}

	return &v1.GetBlockChildrenResponse{
		Blocks: v1blocks,
	}, nil
}

func (a *grpcApi) GetBlockDescendants(ctx context.Context, req *v1.GetBlockDescendantsRequest) (*v1.GetBlockDescendantsResponse, error) {
	logrus.Infof("Getting descendant blocks for block: %s", req.GetBlockId())
	var err error
	blockID, err := uuid.Parse(req.GetBlockId())
	if err != nil {
		return nil, err
	}
	var spaceID uuid.UUID

	if req.SpaceId == nil {
		// get space id from block id
		sid, err := a.api.GetBlockSpaceID(blockID)
		if err != nil {
			return nil, err
		}
		spaceID = *sid
	} else {
		sid, err := uuid.Parse(req.GetSpaceId())
		if err != nil {
			return nil, err
		}
		spaceID = sid
	}

	blocks, err := a.api.GetDescendantBlocks(spaceID, blockID)
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

func (a *grpcApi) GetBlockPage(ctx context.Context, request *v1.GetBlockPageRequest) (*v1.GetBlockPageResponse, error) {
	//TODO implement me
	panic("implement me")
}
