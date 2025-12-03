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

	pubChan, queue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.Durable,
	)
	if err != nil {
		log.Fatalf("problem declaring and binding queue: %v", err)
	}
	fmt.Printf("%v queue declared and bound!\n", queue.Name)

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
			fmt.Printf("Published pause to %v queue\n", queue.Name)
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
			fmt.Printf("Published resume to %v queue\n", queue.Name)
		case "quit":
			fmt.Println("Peril server shutting down...")
			return
		default:
			fmt.Println("command not found")
		}
	}
}
