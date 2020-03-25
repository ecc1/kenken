package main

import (
	"github.com/ecc1/kenken"
)

var (
	puzzle *kenken.Puzzle
	title  string
)

func main() {
	puzzle, title = kenken.ReadPuzzle()
	initGame()
	initUI()
	runUI()
}
