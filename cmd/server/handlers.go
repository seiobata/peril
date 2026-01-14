package main

import (
	"fmt"

	"github.com/seiobata/peril/internal/gamelogic"
	"github.com/seiobata/peril/internal/pubsub"
	"github.com/seiobata/peril/internal/routing"
)

func handlerGameLog() func(routing.GameLog) pubsub.AckType {
	return func(log routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")
		if err := gamelogic.WriteLog(log); err != nil {
			fmt.Println(err)
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
