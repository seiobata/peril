package pubsub

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("problem creating channel: %v", err)
	}

	queue, err := ch.QueueDeclare(
		queueName,
		queueType == Durable,
		queueType != Durable,
		queueType != Durable,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("problem creating queue: %v", err)
	}

	err = ch.QueueBind(
		queue.Name,
		key,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("problem binding queue: %v", err)
	}
	return ch, queue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("could not declare and bind queue: %v", err)
	}

	deliveryCh, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("problem creating consume channel: %v", err)
	}

	go func() {
		defer ch.Close()
		for msg := range deliveryCh {
			var body T
			if err := json.Unmarshal(msg.Body, &body); err != nil {
				log.Printf("could not unmarshal message body: %v", err)
				continue
			}
			acknowledged := handler(body)
			switch acknowledged {
			case Ack:
				msg.Ack(false)
				fmt.Println("Message successfully acknowledged!")
			case NackRequeue:
				msg.Nack(false, true)
				fmt.Println("Message unsuccessful and requeued")
			case NackDiscard:
				msg.Nack(false, false)
				fmt.Println("Message unsuccessful and discarded")
			}
		}
	}()
	return nil
}
