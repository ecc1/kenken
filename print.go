package kenken

import (
	"fmt"
	"io"
)

func (k *KenKen) PrintAnswer(w io.Writer) {
	size := k.Size()

	// Top border
	fmt.Fprint(w, "┏━━━")
	for x := 1; x < size; x++ {
		if k.Vertical[0][x-1] {
			fmt.Fprint(w, "┳━━━")
		} else {
			fmt.Fprint(w, "━━━━")
		}
	}
	fmt.Fprint(w, "┓\n")

	for y := 0; y < size; y++ {
		if y != 0 {
			// Interior border
			if k.Horizontal[0][y-1] {
				fmt.Fprint(w, "┣━━━")
			} else {
				fmt.Fprint(w, "┃   ")
			}
			for x := 1; x < size; x++ {
				fmt.Fprint(w, k.answerCrossing(x, y))
			}
			if k.Horizontal[size-1][y-1] {
				fmt.Fprint(w, "┫\n")
			} else {
				fmt.Fprint(w, "┃\n")
			}
		}

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
			fmt.Fprintf(w, " %d ", k.Answer[y][x])
		}
		fmt.Fprint(w, "┃\n")
	}

	// Bottom border
	fmt.Fprint(w, "┗━━━")
	for x := 1; x < size; x++ {
		if k.Vertical[size-1][x-1] {
			fmt.Fprint(w, "┻━━━")
		} else {
			fmt.Fprint(w, "━━━━")
		}
	}
	fmt.Fprint(w, "┛\n")
}

func (k *KenKen) crossingIndex(x, y int) int {
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

var crossing0 = []string{
	" ", "╸", "╺", "━", "╹", "┛", "┗", "┻",
	"╻", "┓", "┏", "┳", "┃", "┫", "┣", "╋",
}

func (k *KenKen) answerCrossing(x, y int) string {
	c := crossing0[k.crossingIndex(x, y)]
	if k.Horizontal[x][y-1] {
		return fmt.Sprintf("%s━━━", c)
	} else {
		return fmt.Sprintf("%s   ", c)
	}
}

func (k *KenKen) PrintPuzzle(w io.Writer) {
	size := len(k.Answer)

	// Top border
	fmt.Fprint(w, "┏━━━━━━━")
	for x := 1; x < size; x++ {
		if k.Vertical[0][x-1] {
			fmt.Fprint(w, "┳━━━━━━━")
		} else {
			fmt.Fprint(w, "┯━━━━━━━")
		}
	}
	fmt.Fprint(w, "┓\n")

	for y := 0; y < size; y++ {
		if y != 0 {
			// Interior border
			if k.Horizontal[0][y-1] {
				fmt.Fprint(w, "┣━━━━━━━")
			} else {
				fmt.Fprint(w, "┠───────")
			}
			for x := 1; x < size; x++ {
				fmt.Fprint(w, k.puzzleCrossing(x, y))
			}
			if k.Horizontal[size-1][y-1] {
				fmt.Fprint(w, "┫\n")
			} else {
				fmt.Fprint(w, "┨\n")
			}
		}

		// Clue row
		for x := 0; x < size; x++ {
			if x == 0 || k.Vertical[y][x-1] {
				fmt.Fprint(w, "┃")
			} else {
				fmt.Fprint(w, "│")
			}
			switch k.Operation[y][x] {
			case None, Given:
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
					fmt.Fprint(w, "│")
				}
			}
			switch k.Operation[y][x] {
			case Given:
				fmt.Fprintf(w, "   %d   ", k.Answer[y][x])
			default:
				fmt.Fprint(w, "       ")
			}
		}
		fmt.Fprint(w, "┃\n")

		for x := 0; x < size; x++ {
			if x == 0 || k.Vertical[y][x-1] {
				fmt.Fprint(w, "┃       ")
			} else {
				fmt.Fprint(w, "│       ")
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
			fmt.Fprint(w, "┷━━━━━━━")
		}
	}
	fmt.Fprint(w, "┛\n")
}

var crossing1 = []string{
	"┼", "┽", "┾", "┿", "╀", "╃", "╄", "╇",
	"╁", "╅", "╆", "╈", "╂", "╉", "╊", "╋",
}

func (k *KenKen) puzzleCrossing(x, y int) string {
	c := crossing1[k.crossingIndex(x, y)]
	if k.Horizontal[x][y-1] {
		return fmt.Sprintf("%s━━━━━━━", c)
	} else {
		return fmt.Sprintf("%s───────", c)
	}
}

func clueString(clue int, op Operation) string {
	return fmt.Sprintf("%d%s", clue, op.Symbol())
}
