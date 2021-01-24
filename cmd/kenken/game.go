package main

import (
	"fmt"

	"github.com/ecc1/kenken"
)

const (
	autoPromote = true
)

var (
	size        int
	cells       [][]int
	notes       [][][]bool
	noteStrings [][]string
)

func initGame() {
	size = puzzle.Size()
	cells = make([][]int, size)
	notes = make([][][]bool, size)
	noteStrings = make([][]string, size)
	for y := 0; y < size; y++ {
		cells[y] = make([]int, size)
		notes[y] = make([][]bool, size)
		noteStrings[y] = make([]string, size)
		for x := 0; x < size; x++ {
			if isConstant(x, y) {
				cells[y][x] = puzzle.Answer[y][x]
			}
			notes[y][x] = make([]bool, size)
		}
	}
}

func clueString(x, y int) string {
	op := puzzle.Operation[y][x]
	switch op {
	case kenken.None, kenken.Given:
		return ""
	default:
		return fmt.Sprintf("%d %s", puzzle.Clue[y][x], op.Symbol())
	}
}

func addNotes(x, y int) {
	assertNonConstant(x, y)
	for i := 1; i <= size; i++ {
		if cells[y][x] == 0 && !inRowOrColumn(x, y, i) {
			notes[y][x][i-1] = true
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
	cells[y][x] = n
	clearNotes(x, y)
	removeOtherNotes(x, y, n)
	redraw(x, y)
	done, correct := gameStatus()
	if done {
		if correct {
			winnerWinner()
		} else {
			tryAgain()
		}
	}
}

func clearNotes(x, y int) {
	assertNonConstant(x, y)
	for i := 1; i <= size; i++ {
		notes[y][x][i-1] = false
	}
	updateNoteLabel(x, y, false)
}

func clearAll(x, y int) {
	assertNonConstant(x, y)
	cells[y][x] = 0
	clearNotes(x, y)
}

func removeOtherNotes(x, y, n int) {
	assertNonConstant(x, y)
	for i := 1; i <= size; i++ {
		removeNote(i-1, y, n)
		removeNote(x, i-1, n)
	}
}

// For simplicity this can be called on constant cells too.
func removeNote(x, y, n int) {
	if !isConstant(x, y) {
		notes[y][x][n-1] = false
		updateNoteLabel(x, y, true)
	}
}

func updateNote(x, y, n int) {
	assertNonConstant(x, y)
	// Don't allow note if the cell is already set,
	// or if that value is already present in the row or column.
	if cells[y][x] != 0 || inRowOrColumn(x, y, n) {
		return
	}
	notes[y][x][n-1] = !notes[y][x][n-1]
	updateNoteLabel(x, y, false)
	redraw(x, y)
}

func inRowOrColumn(x, y, n int) bool {
	for i := 1; i <= size; i++ {
		if cells[y][i-1] == n || cells[i-1][x] == n {
			return true
		}
	}
	return false
}

func updateNoteLabel(x, y int, promote bool) {
	assertNonConstant(x, y)
	noteStr := ""
	p := 0
	count := 0
	for i := 1; i <= size; i++ {
		if notes[y][x][i-1] {
			noteStr += fmt.Sprintf("%d", i)
			p = i
			count++
		}
	}
	if autoPromote && count == 1 && promote {
		updateCell(x, y, p)
	} else {
		noteStrings[y][x] = noteStr
	}
	redraw(x, y)
}

func gameStatus() (complete bool, correct bool) {
	correct = true
	for j := 0; j < size; j++ {
		for i := 0; i < size; i++ {
			if cells[j][i] == 0 {
				return
			}
			if cells[j][i] != puzzle.Answer[j][i] {
				correct = false
			}
		}
	}
	complete = true
	return
}

func restartGame() {
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if !isConstant(x, y) {
				clearAll(x, y)
			}
		}
	}
}
