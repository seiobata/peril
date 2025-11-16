package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"
)

const connection_string = "amqp://guest:guest@localhost:5672/"

func main() {
	fmt.Println("Starting Peril server...")
	conn, err := amqp.Dial(connection_string)
	if err != nil {
		log.Fatalf("problem connecting to Peril server: %v", err)
	}
	defer conn.Close()
	fmt.Println("Connected to Peril server.")

	// waiting for signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	fmt.Println("Shutting down Peril server...")
}
