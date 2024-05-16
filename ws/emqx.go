package ws

import (
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type EmqxClient struct {
	client mqtt.Client
}

var _ Client = (*EmqxClient)(nil)

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

// NewEmqxClient creates a new EmqxClient
func NewEmqxClient() Client {
	mqtt.DEBUG = log.New(os.Stdout, "", 0)
	mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker("tcp://broker.emqx.io:1883").SetClientID("emqx_test_client")

	opts.SetKeepAlive(60 * time.Second)
	// Set the message callback handler
	opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)

	c := mqtt.NewClient(opts)

	return EmqxClient{client: c}
}

func (e EmqxClient) Connect() error {
	if token := e.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (e EmqxClient) Disconnect() error {
	e.client.Disconnect(250)

	return nil
}

func (e EmqxClient) Publish(topic string, msg string) error {
	token := e.client.Publish(topic, 0, false, msg)
	token.Wait()

	return nil
}

func (e EmqxClient) Subscribe(topic string, cb MessageHandler) (PubSub, error) {
	receiver := func(client mqtt.Client, msg mqtt.Message) {
		data := Message{
			Topic:   msg.Topic(),
			Payload: msg.Payload(),
		}
		cb(e, data)
	}

	if token := e.client.Subscribe(topic, 0, receiver); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return EmqxPubSub{
		topic:  topic,
		client: e.client,
	}, nil
}

type EmqxPubSub struct {
	client mqtt.Client
	topic  string
}

func (e EmqxPubSub) Publish(msg string) error {
	token := e.client.Publish(e.topic, 0, false, msg)
	token.Wait()

	return nil
}

func (e EmqxPubSub) Unsubscribe() error {
	if token := e.client.Unsubscribe(e.topic); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}
