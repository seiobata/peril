package main

import (
	"fmt"
	"log"
	"time"

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

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("problem opening channel: %v", err)
	}

	gs := gamelogic.NewGameState(username)

	// create subscription for pause queue
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
	// create subscription for move queue
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix+".*",
		pubsub.Transient,
		handlerArmyMove(gs, ch),
	)
	if err != nil {
		log.Fatalf("could not create army move subscribe channel: %v", err)
	}

	// create subscription for war queue
	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix, // all clients share the same queue
		routing.WarRecognitionsPrefix+".*",
		pubsub.Durable,
		handlerMakeWar(gs, ch),
	)
	if err != nil {
		log.Fatalf("could not create war subscribe channel: %v", err)
	}

	// start REPL; listen for player commands
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
			move, err := gs.CommandMove(words)
			if err != nil {
				fmt.Printf("problem executing \"move\" command: %v\n", err)
				continue
			}
			err = pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+username,
				move,
			)
			if err != nil {
				fmt.Printf("problem publishing move command: %v\n", err)
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

func publishGameLog(ch *amqp.Channel, msg, username string) error {
	return pubsub.PublishGob(
		ch,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+username,
		routing.GameLog{
			CurrentTime: time.Now(),
			Message:     msg,
			Username:    username,
		},
	)
}
