package messaging

import (
	"encoding/json"
	"time"

	"github.com/fdemchenko/exchanger/cmd/customers/internal/data"
	"github.com/fdemchenko/exchanger/internal/communication"
	"github.com/fdemchenko/exchanger/internal/communication/customers"
	"github.com/fdemchenko/exchanger/internal/communication/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type customerCreationConsumer struct {
	channel       *amqp.Channel
	customersRepo *data.CustomerPostgreSQLRepository
	producer      *rabbitmq.GenericProducer
}

func NewCustomerCreationConsumer(
	channel *amqp.Channel,
	customersRepo *data.CustomerPostgreSQLRepository,
	producer *rabbitmq.GenericProducer,
) *customerCreationConsumer {
	return &customerCreationConsumer{
		channel:       channel,
		customersRepo: customersRepo,
		producer:      producer,
	}
}

func (ccc *customerCreationConsumer) StartListening() error {
	var startingError error
	go func() {
		deliveries, err := ccc.channel.Consume(
			customers.CreateCustomerRequestQueue,
			"",
			false, false, false, false, nil,
		)
		if err != nil {
			startingError = err
			return
		}

		for delivery := range deliveries {
			if err := ccc.handleDelivery(delivery); err != nil {
				log.Error().Err(err).Send()
			}
			if err := delivery.Ack(false); err != nil {
				log.Error().Err(err).Send()
			}
		}
	}()

	return startingError
}

func (ccc *customerCreationConsumer) handleDelivery(delivery amqp.Delivery) error {
	msg := communication.Message[json.RawMessage]{}
	err := json.Unmarshal(delivery.Body, &msg)
	if err != nil {
		return err
	}
	switch msg.Type {
	case customers.CreateCustomerRequest:
		request := customers.CreateCustomerRequestPayload{}
		err := json.Unmarshal(msg.Payload, &request)
		if err != nil {
			return err
		}
		return ccc.handleCustomerCreation(request)
	default:
		log.Error().Msg("Invalid message type in customer creation")
	}
	return nil
}

func (ccc *customerCreationConsumer) handleCustomerCreation(request customers.CreateCustomerRequestPayload) error {
	id, err := ccc.customersRepo.Insert(request.Email, request.SubscriptionID)
	var message any
	if err != nil {
		message = communication.Message[customers.CustomerCreationFailedPayload]{
			MessageHeader: communication.MessageHeader{Type: customers.CustomerCreationFailed, Timestamp: time.Now()},
			Payload:       customers.CustomerCreationFailedPayload{Error: err.Error(), SubscriptionID: request.SubscriptionID},
		}
	} else {
		message = communication.Message[customers.CustomerCreatedPayload]{
			MessageHeader: communication.MessageHeader{Type: customers.CustomerCreated, Timestamp: time.Now()},
			Payload:       customers.CustomerCreatedPayload{ID: id},
		}
	}
	return ccc.producer.SendMessage(message, customers.CreateCustomerResponseQueue)
}
