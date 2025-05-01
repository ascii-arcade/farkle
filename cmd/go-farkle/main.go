package main

import (
	"fmt"
	"os"

	"github.com/kthibodeaux/go-farkle/internal/tui"
)

func main() {
	playerNames := os.Args[1:]
	if len(playerNames) == 0 {
		fmt.Println("Usage: go-farkle player1 player2 ...")
		return
	}

	tui.Run(playerNames)
}
