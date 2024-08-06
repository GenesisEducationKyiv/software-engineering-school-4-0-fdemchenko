package messaging

import (
	"encoding/json"

	"github.com/fdemchenko/exchanger/internal/communication"
	"github.com/fdemchenko/exchanger/internal/communication/customers"
	"github.com/fdemchenko/exchanger/internal/repositories"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type customerCreationSAGAConsumer struct {
	channel                *amqp.Channel
	subscriptionRepository *repositories.PostgresSubscriptionRepository
}

func NewCustomerCreationSAGAConsumer(
	channel *amqp.Channel,
	subscriptionRepository *repositories.PostgresSubscriptionRepository,
) *customerCreationSAGAConsumer {
	return &customerCreationSAGAConsumer{
		channel:                channel,
		subscriptionRepository: subscriptionRepository,
	}
}

func (ccsc *customerCreationSAGAConsumer) StartListening() error {
	var startingError error
	go func() {
		deliveries, err := ccsc.channel.Consume(
			customers.CreateCustomerResponseQueue,
			"",
			false, false, false, false, nil,
		)
		if err != nil {
			startingError = err
			return
		}

		for delivery := range deliveries {
			if err := ccsc.handleDelivery(delivery); err != nil {
				log.Error().Err(err).Send()
			}
			if err := delivery.Ack(false); err != nil {
				log.Error().Err(err).Send()
			}
		}
	}()

	return startingError
}

func (ccsc *customerCreationSAGAConsumer) handleDelivery(delivery amqp.Delivery) error {
	msg := communication.Message[json.RawMessage]{}
	err := json.Unmarshal(delivery.Body, &msg)
	if err != nil {
		return err
	}
	switch msg.Type {
	case customers.CustomerCreated:
		request := customers.CustomerCreatedPayload{}
		err := json.Unmarshal(msg.Payload, &request)
		if err != nil {
			return err
		}

		log.Info().Int("customer_id", request.ID).Msg("Customer created")
	case customers.CustomerCreationFailed:
		request := customers.CustomerCreationFailedPayload{}
		err := json.Unmarshal(msg.Payload, &request)
		if err != nil {
			return err
		}
		// compensate
		err = ccsc.subscriptionRepository.DeleteByID(request.SubscriptionID)
		log.Error().Int("subscription_id", request.SubscriptionID).
			Msg("Failed to create customer, running compensate transaction")
		if err != nil {
			log.Error().Err(err).Msg("Compensation transaction failed")
		}
	default:
		log.Error().Msg("Invalid message type in customer creation")
	}
	return nil
}
