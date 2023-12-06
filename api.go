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

func (a Api) ApplyTransactions(ctx context.Context, req *v1.ApplyTransactionRequest) (*v1.ApplyTransactionResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (a Api) CreateSpace(ctx context.Context, req *v1.CreateSpaceRequest) (*v1.CreateSpaceResponse, error) {
	//TODO implement me
	panic("implement me")
}
