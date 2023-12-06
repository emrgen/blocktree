package blocktree

import (
	"context"
	v1 "github.com/emrgen/blocktree/apis/v1"
)

var (
	_ v1.BlocktreeServer = (*Api)(nil)
)

type Api struct {
	v1.BlocktreeServer
	store Store
}

func (s *Api) CreateTransaction(ctx context.Context, transaction *v1.Transaction) (*v1.Transaction, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Api) CreateSpace(ctx context.Context, request *v1.CreateSpaceRequest) (*v1.CreateSpaceResponse, error) {
	//TODO implement me
	panic("implement me")
}
