package main

import (
	"io"
	"flag"
	"fmt"
	"os"

	"github.com/ecc1/kenken"
)

var (
	showAnswers = flag.Bool("s", false, "display solution")
)

func main() {
	k, _ := kenken.ReadPuzzle()
	printPuzzle(k, os.Stdout)
}

// PrintPuzzle prints a puzzle to w.
func printPuzzle(k *kenken.Puzzle, w io.Writer) {
	size := len(k.Answer)

	// Top border
	fmt.Fprint(w, "┏━━━━━━━")
	for x := 1; x < size; x++ {
		if k.Vertical[0][x-1] {
			fmt.Fprint(w, "┳━━━━━━━")
		} else {
			fmt.Fprint(w, "━━━━━━━━")
		}
	}
	fmt.Fprint(w, "┓\n")

	for y := 0; y < size; y++ {
		if y != 0 {
			// Interior border
			if k.Horizontal[0][y-1] {
				fmt.Fprint(w, "┣━━━━━━━")
			} else {
				fmt.Fprint(w, "┃       ")
			}
			for x := 1; x < size; x++ {
				fmt.Fprint(w, puzzleCrossing(k, x, y))
			}
			if k.Horizontal[size-1][y-1] {
				fmt.Fprint(w, "┫\n")
			} else {
				fmt.Fprint(w, "┃\n")
			}
		}

		// Clue row
		for x := 0; x < size; x++ {
			if x == 0 || k.Vertical[y][x-1] {
				fmt.Fprint(w, "┃")
			} else {
				fmt.Fprint(w, " ")
			}
			switch k.Operation[y][x] {
			case kenken.None, kenken.Given:
				fmt.Fprint(w, "       ")
			default:
				fmt.Fprintf(w, "%-7s", clueString(k.Clue[y][x], k.Operation[y][x]))
			}
		}
		fmt.Fprint(w, "┃\n")

		// Cell contents
		fmt.Fprint(w, "┃")
		for x := 0; x < size; x++ {
			if x != 0 {
				if k.Vertical[y][x-1] {
					fmt.Fprint(w, "┃")
				} else {
					fmt.Fprint(w, " ")
				}
			}
			if *showAnswers || k.Operation[y][x] == kenken.Given {
				fmt.Fprintf(w, "   %d   ", k.Answer[y][x])
			} else {
				fmt.Fprint(w, "       ")
			}
		}
		fmt.Fprint(w, "┃\n")

		for x := 0; x < size; x++ {
			if x == 0 || k.Vertical[y][x-1] {
				fmt.Fprint(w, "┃       ")
			} else {
				fmt.Fprint(w, "        ")
			}
		}
		fmt.Fprint(w, "┃\n")
	}

	// Bottom border
	fmt.Fprint(w, "┗━━━━━━━")
	for x := 1; x < size; x++ {
		if k.Vertical[size-1][x-1] {
			fmt.Fprint(w, "┻━━━━━━━")
		} else {
			fmt.Fprint(w, "━━━━━━━━")
		}
	}
	fmt.Fprint(w, "┛\n")
}

func crossingIndex(k *kenken.Puzzle, x, y int) int {
	index := 0
	if k.Horizontal[x-1][y-1] {
		index |= 1 << 0
	}
	if k.Horizontal[x][y-1] {
		index |= 1 << 1
	}
	if k.Vertical[y-1][x-1] {
		index |= 1 << 2
	}
	if k.Vertical[y][x-1] {
		index |= 1 << 3
	}
	return index
}

var crossing = []string{
	" ", "╸", "╺", "━", "╹", "┛", "┗", "┻",
	"╻", "┓", "┏", "┳", "┃", "┫", "┣", "╋",
}

func puzzleCrossing(k *kenken.Puzzle, x, y int) string {
	c := crossing[crossingIndex(k, x, y)]
	if k.Horizontal[x][y-1] {
		return fmt.Sprintf("%s━━━━━━━", c)
	}
	return fmt.Sprintf("%s       ", c)
}

func clueString(clue int, op kenken.Operation) string {
	return fmt.Sprintf("%d%s", clue, op.Symbol())
}
