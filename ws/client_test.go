package ws

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Set up before tests
	m.Run()
	// Tear down after tests
}

func TestEmqxClient_Subscribe(t *testing.T) {
	testClientSubscription(t, NewEmqxClient())
	testClientSubscription(t, NewRedisClient())
}

func testClientSubscription(t *testing.T, client Client) {
	err := client.Connect()
	assert.NoError(t, err)

	received := make(chan []byte)

	subscribe, err := client.Subscribe("test/1", func(client Client, msg Message) {
		received <- msg.Payload
	})
	assert.NoError(t, err)

	err = subscribe.Publish("Hello, World!")
	assert.NoError(t, err)

	var msg []byte
	select {
	case msg = <-received:
		t.Logf("Received message: %s", msg)
		assert.Equal(t, "Hello, World!", string(msg))
		break
	case <-time.After(5 * time.Second):
		t.Error("Timeout")
	}
}
