package main

import (
	"os"

	"github.com/ecc1/kenken"
)

func main() {
	k, _ := kenken.ReadPuzzle()
	k.PrintAnswer(os.Stdout)
}
