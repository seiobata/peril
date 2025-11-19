package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/seiobata/peril/internal/pubsub"
	"github.com/seiobata/peril/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

const connection_string = "amqp://guest:guest@localhost:5672/"

func main() {
	fmt.Println("Starting Peril server...")
	conn, err := amqp.Dial(connection_string)
	if err != nil {
		log.Fatalf("problem connecting to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril server connected to RabbitMQ!")

	pubChan, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel could not be opened: %v", err)
	}

	err = pubsub.PublishJSON(
		pubChan,
		routing.ExchangePerilDirect,
		routing.PauseKey,
		routing.PlayingState{
			IsPaused: true,
		},
	)
	if err != nil {
		log.Printf("problem publishing message: %v", err)
	}

	// waiting for signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	fmt.Println("Shutting down Peril server...")
}
