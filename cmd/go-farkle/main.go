package main

import (
	"fmt"
	"os"

	"github.com/kthibodeaux/go-farkle/internal/tui"
)

func main() {
	playerNames := os.Args[1:]
	if len(playerNames) == 0 {
		fmt.Println("Usage: program item1 item2 item3 ...")
		return
	}

	tui.Run(playerNames)
}
