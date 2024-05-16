package ws

import (
	"errors"

	"github.com/sirupsen/logrus"

	"github.com/go-redis/redis"
)

var _ Client = (*RedisClient)(nil)

type RedisClient struct {
	redis *redis.Client
}

// NewRedisClient creates a new RedisClient
func NewRedisClient() Client {
	return RedisClient{
		redis: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
	}
}

func (r RedisClient) Publish(topic string, msg string) error {
	_, err := r.redis.Publish(topic, msg).Result()
	if err != nil {
		return err
	}

	return nil
}

func (r RedisClient) Subscribe(topic string, cb MessageHandler) (PubSub, error) {
	sub := r.redis.Subscribe(topic)
	payload, err := sub.Receive()
	if err != nil {
		return nil, err
	}

	var msg []byte
	switch payload.(type) {
	case []byte:
		msg = payload.([]byte)
	case string:
		msg = []byte(payload.(string))
	default:
		logrus.Printf("invalid payload type: %T", payload)
		return nil, errors.New("invalid payload type")
	}

	cb(r, Message{
		Topic:   topic,
		Payload: msg,
	})

	return RedisPubSub{
		topic:  topic,
		client: r.redis,
		pubsub: sub,
	}, nil
}

func (r RedisClient) Connect() error {
	return nil
}

func (r RedisClient) Disconnect() error {
	return r.redis.Close()
}

type RedisPubSub struct {
	topic  string
	client *redis.Client
	pubsub *redis.PubSub
}

func (r RedisPubSub) Publish(msg string) error {
	_, err := r.client.Publish(r.topic, msg).Result()
	if err != nil {
		return err
	}

	return nil
}

func (r RedisPubSub) Unsubscribe() error {
	err := r.pubsub.Unsubscribe(r.topic)
	if err != nil {
		return err
	}

	return nil
}
