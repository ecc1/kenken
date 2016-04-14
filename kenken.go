/*
Package kenken provides functions to read KenKen puzzles
in the format used by http://app.kenkenpuzzle.com/kenken/puzzles/
*/
package kenken

type KenKen struct {
	Answer    [][]int
	Clue      [][]int
	Operation [][]Operation

	// N rows of N-1 columns
	// Vertical[y][x-1] == true means there is a heavy vertical line
	// between (x-1, y) and (x, y)
	Vertical [][]bool

	// N rows of N-1 columns
	// Horizontal[x][y-1] == true means there is a heavy horizontal line
	// between (x, y-1) and (x, y) (note transpose)
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

func (k *KenKen) Size() int {
	return len(k.Answer)
}

var ops = []string{"?", "?", "+", "−", "×", "∕"}

func (op Operation) Symbol() string {
	return ops[op]
}
