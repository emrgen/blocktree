package blocktree

import (
	"context"
	"testing"
	"time"

	v1 "github.com/emrgen/blocktree/apis/v1"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func createConnection(port string) (*grpc.ClientConn, error) {
	return grpc.Dial(port, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func TestApi_Start(t *testing.T) {
	server := NewServer(NewMemStore(), &Config{
		GrpcPort: 4100,
		HttpPort: 4200,
	})

	go func() {
		err := server.Start()
		if err != nil {
			panic(err)
		}
	}()

	time.Sleep(1 * time.Second)

	conn, err := createConnection(":4100")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := v1.NewBlocktreeClient(conn)

	spaceID := uuid.New().String()

	logrus.Infof("Creating a space: %s", spaceID)
	res, err := client.CreateSpace(context.Background(), &v1.CreateSpaceRequest{
		SpaceId: spaceID,
		Name:    "space-name",
	})

	assert.NoError(t, err)
	assert.Equal(t, spaceID, res.SpaceId)

	getBlockRes, err := client.GetBlock(context.TODO(), &v1.GetBlockRequest{
		SpaceId: &spaceID,
		BlockId: spaceID,
	})
	assert.NoError(t, err)

	assert.Equal(t, getBlockRes.Block.BlockId, spaceID)

}
