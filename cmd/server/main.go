package main

import (
	"fmt"
	"log"

	"github.com/seiobata/peril/internal/gamelogic"
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

	gamelogic.PrintServerHelp()
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "pause":
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
		case "resume":
			err = pubsub.PublishJSON(
				pubChan,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)
			if err != nil {
				log.Printf("problem publishing message: %v", err)
			}
		case "quit":
			fmt.Println("Peril server shutting down...")
			return
		default:
			fmt.Println("command not found")
		}
	}
}
