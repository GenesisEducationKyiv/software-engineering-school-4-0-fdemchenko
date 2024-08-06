package rabbitmq

import (
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

const PublishingContentType = "application/json"

type GenericProducer struct {
	channel *amqp.Channel
}

func NewGenericProducer(
	channel *amqp.Channel,
) *GenericProducer {
	return &GenericProducer{
		channel: channel,
	}
}

func (gp *GenericProducer) SendMessage(msg any, queue string) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return gp.channel.Publish("", queue, false, false, amqp.Publishing{
		ContentType: PublishingContentType,
		Body:        body,
	})
}
