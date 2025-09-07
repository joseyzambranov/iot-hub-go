package ports

type MessageHandler interface {
	HandleMessage(topic string, payload []byte) error
}