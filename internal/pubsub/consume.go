package pubsub

import (
	"bytes"
	"encoding/gob"
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
		amqp.Table{
			"x-dead-letter-exchange": "peril_dlx", // send discarded messages to peril_dlx
		},
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
	return subscribe(
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		jsonUnmarshal[T],
	)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	return subscribe(
		conn,
		exchange,
		queueName,
		key,
		queueType,
		handler,
		gobUnmarshal[T],
	)
}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("could not declare and bind queue: %v", err)
	}

	if err = ch.Qos(10, 0, false); err != nil {
		return fmt.Errorf("could not apply prefetch count: %v", err)
	}

	msgs, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("problem creating consume channel: %v", err)
	}

	go func() {
		defer ch.Close()
		for msg := range msgs {
			body, err := unmarshaller(msg.Body)
			if err != nil {
				log.Printf("could not unmarshal message body: %v", err)
				continue
			}
			switch handler(body) {
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

func jsonUnmarshal[T any](dat []byte) (T, error) {
	var body T
	if err := json.Unmarshal(dat, &body); err != nil {
		fmt.Printf("problem unmarshaling JSON: %v", err)
		return body, err
	}
	return body, nil
}

func gobUnmarshal[T any](dat []byte) (T, error) {
	buf := bytes.NewBuffer(dat)
	dec := gob.NewDecoder(buf)
	var body T
	if err := dec.Decode(&body); err != nil {
		fmt.Printf("problem unmarshaling Gob: %v", err)
		return body, err
	}
	return body, nil
}
