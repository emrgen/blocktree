package cmd

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"strings"
)

var (
	nilID = "00000000-0000-0000-0000-000000000000"
)

func createConnection(port string) (*grpc.ClientConn, error) {
	return grpc.Dial(port, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func sanitizeID(id string) string {
	id = strings.ToLower(id)

	if len(id) <= 36 {
		idLen := len(id)
		return nilID[:36-idLen] + id
	}

	if len(id) > 36 {
		return id[:36]
	}

	return id
}
