package main

import (
	"fmt"

	"github.com/seiobata/peril/internal/gamelogic"
	"github.com/seiobata/peril/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}
