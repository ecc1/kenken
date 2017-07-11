/*
Package kenken provides functions to read KenKen puzzles
in the format used by http://app.kenkenpuzzle.com/kenken/puzzles/
*/
package kenken

// Puzzle represents a KenKen puzzle.
type Puzzle struct {
	Answer    [][]int
	Clue      [][]int
	Operation [][]Operation

	// N rows of N-1 columns
	// Vertical[y][x] is true if there is a heavy vertical line
	// between (x, y) and (x+1, y)
	Vertical [][]bool

	// N rows of N-1 columns
	// Horizontal[x][y] is true if there is a heavy horizontal line
	// between (x, y) and (x, y+1) (note transpose)
	Horizontal [][]bool
}

// Operation represents an arithmetic operation for a cage.
type Operation byte

// Cage operations.
const (
	None Operation = iota
	Given
	Sum
	Difference
	Product
	Quotient
)

// Size returns the puzzle size.
func (k *Puzzle) Size() int {
	return len(k.Answer)
}

var ops = []string{"?", "?", "+", "−", "×", "∕"}

// Symbol returns the character corresponding to an operation.
func (op Operation) Symbol() string {
	return ops[op]
}
