package ws

import (
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
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

func (r RedisClient) Publish(topic string, msg []byte) error {
	_, err := r.redis.Publish(topic, msg).Result()
	if err != nil {
		return err
	}

	return nil
}

func (r RedisClient) Subscribe(topic string, cb MessageHandler) (PubSub, error) {
	sub := r.redis.Subscribe(topic)

	channel := sub.Channel()

	go func() {
		for msg := range channel {
			logrus.Print("Received message: ", msg.Payload)
			cb(r, Message{
				Topic:   msg.Channel,
				Payload: []byte(msg.Payload),
			})
		}
	}()

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
