package cmd

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func createConnection(port string) (*grpc.ClientConn, error) {
	return grpc.Dial(port, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
