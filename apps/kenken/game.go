package main

import (
	"fmt"

	"github.com/ecc1/kenken"
)

var game struct {
	k    *kenken.Puzzle
	size int
	cell [][]int
	note [][][]bool
}

var (
	autoPromote = true
)

func initGame() {
	game.size = puzzle.Size()
	game.cell = make([][]int, game.size)
	game.note = make([][][]bool, game.size)
	for y := 0; y < game.size; y++ {
		game.cell[y] = make([]int, game.size)
		game.note[y] = make([][]bool, game.size)
		for x := 0; x < game.size; x++ {
			if isConstant(x, y) {
				game.cell[y][x] = puzzle.Answer[y][x]
			}
			game.note[y][x] = make([]bool, game.size)
		}
	}
}

func clueString(x, y int) string {
	switch puzzle.Operation[y][x] {
	case kenken.None, kenken.Given:
		return ""
	default:
		return fmt.Sprintf("%d %s", puzzle.Clue[y][x], puzzle.Operation[y][x].Symbol())
	}
}

func addNotes(x, y int) {
	assertNonConstant(x, y)
	for i := 1; i <= game.size; i++ {
		if game.cell[y][x] == 0 && !inRowOrColumn(x, y, i) {
			game.note[y][x][i-1] = true
		}
	}
	updateNoteLabel(x, y, true)
}

func isConstant(x, y int) bool {
	return puzzle.Operation[y][x] == kenken.Given
}

func assertNonConstant(x, y int) {
	if isConstant(x, y) {
		panic(fmt.Sprintf("modifying constant cell (%d, %d)", x, y))
	}
}

func updateCell(x, y, n int) {
	assertNonConstant(x, y)
	game.cell[y][x] = n
	setCellLabel(x, y, n)
	clearNotes(x, y)
	removeOtherNotes(x, y, n)
	if isSolved() {
		winnerWinner()
	}
}

func clearNotes(x, y int) {
	assertNonConstant(x, y)
	for i := 1; i <= game.size; i++ {
		game.note[y][x][i-1] = false
	}
	updateNoteLabel(x, y, false)
}

func clearAll(x, y int) {
	assertNonConstant(x, y)
	game.cell[y][x] = 0
	clearCellLabel(x, y)
	clearNotes(x, y)
}

func removeOtherNotes(x, y, n int) {
	assertNonConstant(x, y)
	for i := 1; i <= game.size; i++ {
		removeNote(i-1, y, n)
		removeNote(x, i-1, n)
	}
}

// For simplicity this can be called on constant cells too.
func removeNote(x, y, n int) {
	if !isConstant(x, y) {
		game.note[y][x][n-1] = false
		updateNoteLabel(x, y, true)
	}
}

func updateNote(x, y, n int) {
	assertNonConstant(x, y)
	// Don't allow note if the cell is already set,
	// or if that value is already present in the row or column.
	if game.cell[y][x] != 0 || inRowOrColumn(x, y, n) {
		return
	}
	game.note[y][x][n-1] = !game.note[y][x][n-1]
	updateNoteLabel(x, y, false)
}

func inRowOrColumn(x, y, n int) bool {
	for i := 1; i <= game.size; i++ {
		if game.cell[y][i-1] == n || game.cell[i-1][x] == n {
			return true
		}
	}
	return false
}

func updateNoteLabel(x, y int, promote bool) {
	assertNonConstant(x, y)
	notes := ""
	p := 0
	count := 0
	for i := 1; i <= game.size; i++ {
		if game.note[y][x][i-1] {
			notes += fmt.Sprintf("%d", i)
			p = i
			count++
		}
	}
	if autoPromote && count == 1 && promote {
		updateCell(x, y, p)
	} else {
		setNoteLabel(x, y, notes)
	}
}

func isSolved() bool {
	for j := 0; j < game.size; j++ {
		for i := 0; i < game.size; i++ {
			if game.cell[j][i] != puzzle.Answer[j][i] {
				return false
			}
		}
	}
	return true
}
