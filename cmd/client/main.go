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
	fmt.Println("Starting Peril client...")
	conn, err := amqp.Dial(connection_string)
	if err != nil {
		log.Fatalf("problem connecting to RabbitMQ: %v", err)
	}
	defer conn.Close()

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("problem retrieving username: %v", err)
	}

	_, queue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		log.Fatalf("could not declare and bind to pause queue: %v", err)
	}
	fmt.Printf("%v queue declared and bound!\n", queue.Name)

	client := gamelogic.NewGameState(username)

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			err = client.CommandSpawn(words)
			if err != nil {
				fmt.Printf("problem executing \"spawn\" command: %v\n", err)
				continue
			}

		case "move":
			_, err := client.CommandMove(words)
			if err != nil {
				fmt.Printf("problem executing \"move\" command: %v\n", err)
				continue
			}
			fmt.Println("Move command successful!")

		case "status":
			client.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			fmt.Println("Spamming is not allowed yet!")

		case "quit":
			gamelogic.PrintQuit()
			return

		default:
			fmt.Printf("Unknown command. Enter \"help\" for list of commands.\n")
		}
	}
}
