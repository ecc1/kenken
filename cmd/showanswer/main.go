package main

import (
	"os"

	"github.com/ecc1/kenken"
	"github.com/ecc1/kenken/text"
)

func main() {
	k, _ := kenken.ReadPuzzle()
	text.PrintAnswer(k, os.Stdout)
}
