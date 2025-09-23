package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType string

const (
	Durable   SimpleQueueType = "durable"
	Transient SimpleQueueType = "transient"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	j, err := json.Marshal(val)
	if err != nil {
		return err
	}
	ch.PublishWithContext(
		context.Background(), exchange, key,
		false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        j,
		})
	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transi"
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	q, err := ch.QueueDeclare(queueName,
		queueType == SimpleQueueType(Durable),
		queueType == SimpleQueueType(Transient),
		queueType == SimpleQueueType(Transient),
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	ch.QueueBind(queueName, key, exchange, false, nil)
	return ch, q, nil
}
