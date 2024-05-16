package ws

import (
	"github.com/nsqio/go-nsq"
)

type NsqClient struct {
	consumer *nsq.Consumer
	producer *nsq.Producer
}

var _ Client = (*NsqClient)(nil)

func (n NsqClient) Publish(topic string, msg string) error {
	return n.producer.Publish(topic, []byte(msg))
}

func (n NsqClient) Subscribe(topic string, cb MessageHandler) (PubSub, error) {
	n.consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		cb(n, Message{
			Topic:   topic,
			Payload: message.Body,
		})

		return nil
	}))

	if err := n.consumer.ConnectToNSQLookupd("localhost:4161"); err != nil {
		return nil, err
	}

	return NsqPubSub{
		topic:  topic,
		client: &n,
	}, nil
}

func (n NsqClient) Connect() error {
	return nil
}

func (n NsqClient) Disconnect() error {
	n.producer.Stop()
	n.consumer.Stop()

	return nil
}

type NsqPubSub struct {
	topic  string
	client *NsqClient
}

func (n NsqPubSub) Publish(msg string) error {
	return n.client.producer.Publish(n.topic, []byte(msg))
}

func (n NsqPubSub) Unsubscribe() error {
	n.client.consumer.Stop()

	return nil
}
