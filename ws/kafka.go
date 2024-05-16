package ws

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaClient struct {
	conn    *kafka.Conn
	closeCh chan struct{}
}

var _ Client = (*KafkaClient)(nil)

// NewKafkaClient creates a new KafkaClient
func NewKafkaClient(topic string, partition int) Client {
	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	return KafkaClient{
		conn:    conn,
		closeCh: make(chan struct{}),
	}
}

func (k KafkaClient) Publish(topic string, msg []byte) error {
	err := k.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return err
	}

	_, err = k.conn.WriteMessages(
		kafka.Message{Value: msg},
	)

	return err
}

func (k KafkaClient) Subscribe(topic string, cb MessageHandler) (PubSub, error) {
	go func() {
		for {
			select {
			case <-k.closeCh:
				return
			default:
				m, err := k.conn.ReadMessage(10e6)
				if err != nil {
					log.Fatal("failed to read message:", err)
				}

				cb(k, Message{
					Topic:   topic,
					Payload: m.Value,
				})
			}
		}
	}()

	return KafkaPubSub{
		client: k,
	}, nil
}

func (k KafkaClient) Connect() error {
	return nil
}

func (k KafkaClient) Disconnect() error {
	close(k.closeCh)
	return k.conn.Close()
}

type KafkaPubSub struct {
	closeCh chan struct{}
	client  KafkaClient
}

func (k KafkaPubSub) Publish(msg string) error {
	return k.client.Publish("", []byte(msg))
}

func (k KafkaPubSub) Unsubscribe() error {
	close(k.client.closeCh)
	return nil
}
