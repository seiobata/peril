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

	gs := gamelogic.NewGameState(username)

	// create new subscribe channel for pause queue
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gs),
	)
	if err != nil {
		log.Fatalf("could not create pause subscribe channel: %v", err)
	}

	// listen for player commands
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			err = gs.CommandSpawn(words)
			if err != nil {
				fmt.Printf("problem executing \"spawn\" command: %v\n", err)
				continue
			}

		case "move":
			_, err := gs.CommandMove(words)
			if err != nil {
				fmt.Printf("problem executing \"move\" command: %v\n", err)
				continue
			}
			fmt.Println("Move command successful!")

		case "status":
			gs.CommandStatus()

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
