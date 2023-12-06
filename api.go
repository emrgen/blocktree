package blocktree

import (
	"context"
	"errors"
	v1 "github.com/emrgen/blocktree/apis/v1"
	"github.com/google/uuid"
)

var (
	_ v1.BlocktreeServer = (*Api)(nil)
)

type Api struct {
	v1.BlocktreeServer
	store Store
}

func NewApi(store Store) *Api {
	return &Api{
		store: store,
	}
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

// ApplyTransactions applies a list of transactions to the blocktree store
func (a *Api) ApplyTransactions(ctx context.Context, req *v1.ApplyTransactionRequest) (*v1.ApplyTransactionResponse, error) {
	txs := req.GetTransactions()
	transactions := make([]*Transaction, len(txs))
	for _, tx := range txs {
		transaction, err := TransactionFromProtoV1(tx)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	res := &v1.ApplyTransactionResponse{
		Transactions: make([]*v1.ApplyTransactionResult, len(transactions)),
	}

	for _, tx := range transactions {
		change, err := tx.Prepare(a.store)
		if err != nil {
			if errors.Is(err, ErrDetectedCycle) || errors.Is(err, ErrCreatesCycle) {
				continue
			}

			return nil, err
		}
		err = a.store.Apply(&tx.SpaceID, change)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}
