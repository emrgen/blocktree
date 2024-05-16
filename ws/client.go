package ws

// Client is the interface that wraps the basic methods to interact with a websocket server.
type Client interface {
	Publish(topic string, msg string) error
	Subscribe(topic string, cb MessageHandler) (PubSub, error)
	Connect() error
	Disconnect() error
}

// Message is the struct that wraps the topic and the payload of a message.
type Message struct {
	Topic   string
	Payload []byte
}

// PubSub is the interface that wraps the basic methods to interact with a subscription to a topic.
type PubSub interface {
	Publish(msg string) error
	Unsubscribe() error
}

// MessageHandler is the type that wraps the function that will be called when a message is received.
type MessageHandler = func(client Client, msg Message)
