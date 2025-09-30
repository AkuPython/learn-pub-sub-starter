package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType string
type AckType int

const (
	Durable   SimpleQueueType = "durable"
	Transient SimpleQueueType = "transient"
)

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
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
		amqp.Table{"x-dead-letter-exchange": "peril_dlx"})
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("issue with queueDeclare - %v", err)
	}
	ch.QueueBind(queueName, key, exchange, false, nil)
	return ch, q, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	ch, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	dS, err := ch.Consume(queueName, "", false, false, false, false, nil)

	if err != nil {
		return err
	}
	go func() {
		for d := range dS {
			var message T
			if err := json.Unmarshal(d.Body, &message); err != nil {
				fmt.Println(err)
				continue
			}
			switch handler(message) {
			case Ack:
				d.Ack(false)
				fmt.Println("Ack")
			case NackDiscard:
				d.Nack(false, false)
				fmt.Println("NackDiscard")
			case NackRequeue:
				d.Nack(false, true)
				fmt.Println("NackRequeue")
			}

		}
	}()
	return nil
}
