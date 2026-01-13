package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	msg, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("error marshalling message: %v", err)
	}

	err = ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msg,
		})
	if err != nil {
		return fmt.Errorf("error publishing message: %v", err)
	}
	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var msg bytes.Buffer
	enc := gob.NewEncoder(&msg)
	err := enc.Encode(val)
	if err != nil {
		return fmt.Errorf("error encoding message: %v", err)
	}

	err = ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/gob",
			Body:        msg.Bytes(),
		},
	)
	if err != nil {
		return fmt.Errorf("error publishing message: %v", err)
	}
	return nil
}
