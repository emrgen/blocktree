package blocktree

import "testing"

func TestServer_Start(t *testing.T) {
	server := NewServer(NewMemStore(), &Config{
		GrpcPort: 4001,
		HttpPort: 4002,
	})

	go func() {
		err := server.Start()
		if err != nil {
			panic(err)
		}
	}()

}
