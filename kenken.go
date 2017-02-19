/*
Package kenken provides functions to read KenKen puzzles
in the format used by http://app.kenkenpuzzle.com/kenken/puzzles/
*/
package kenken

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

type Operation byte

const (
	None Operation = iota
	Given
	Sum
	Difference
	Product
	Quotient
)

func (k *Puzzle) Size() int {
	return len(k.Answer)
}

var ops = []string{"?", "?", "+", "−", "×", "∕"}

func (op Operation) Symbol() string {
	return ops[op]
}
