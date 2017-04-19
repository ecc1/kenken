package main

import (
	"github.com/ecc1/kenken"
)

var (
	puzzle *kenken.Puzzle
	path   string
)

func main() {
	puzzle, path = kenken.ReadPuzzle()
	initGame()
	initUI()
	runUI()
}
