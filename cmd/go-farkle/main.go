package main

import (
	"fmt"
	"math/rand/v2"
	"os"

	"github.com/kthibodeaux/go-farkle/internal/tui"
)

func main() {
	playerNames := os.Args[1:]
	if len(playerNames) == 0 {
		fmt.Println("Usage: go-farkle player1 player2 ...")
		return
	}

	rand.Shuffle(len(playerNames), func(i, j int) {
		playerNames[i], playerNames[j] = playerNames[j], playerNames[i]
	})
	tui.Run(playerNames)
}
