package kenken

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

var debugging = false

func debug(format string, args ...interface{}) {
	if !debugging {
		return
	}
	fmt.Printf(format, args...)
}

// ReadSGT constructs a puzzle from the given SGT puzzle id and solution strings.
func ReadSGT(id string, soln string) (*Puzzle, error) {
	if id[0] < '0' || id[0] > '9' || id[1] != ':' {
		log.Printf("expected size and colon here: %s", id)
		return nil, errMalformed
	}
	size := int(id[0] - '0')
	k := &Puzzle{
		Answer:     makeIntMatrix(size),
		Clue:       makeIntMatrix(size),
		Operation:  makeOpMatrix(size),
		Vertical:   makeBoolMatrix(size),
		Horizontal: makeBoolMatrix(size),
	}
	s := decompressSGTString(id[2:])
	debug("%s\n", s)
	i := readSGTLines(s, k)
	if i < 0 {
		return nil, errMalformed
	}
	err := readSGTClues(s[i:], k)
	if err != nil {
		return nil, err
	}
	err = readSGTSolution(soln, k)
	return k, err
}

// Read the first half of the string to set the Vertical and Horizontal matrices.
// Return the index of the second half of the string, or -1 if malformed.
func readSGTLines(s string, k *Puzzle) int {
	size := k.Size()
	pos := 0
	end := 2 * size * (size - 1)
	i := 0
	for i < len(s) {
		c := s[i]
		i++
		if c == '_' {
			pos++
		} else if 'a' <= c && c <= 'z' {
			pos += int(c-'a') + 2
		} else {
			log.Printf("expected letter or underscore here: %s", s[i-1:])
			return -1
		}
		if pos > end {
			break
		}
		addSGTLine(pos, k)
	}
	if s[i] != ',' {
		log.Printf("expected comma here: %s", s[i:])
		return -1
	}
	return i + 1
}

func addSGTLine(pos int, k *Puzzle) {
	size := k.Size()
	middle := size * (size - 1)
	x := pos % (size - 1)
	y := pos / (size - 1)
	if x == 0 {
		x = size - 1
		y--
	}
	if pos >= middle && y != size-1 {
		y -= size
	}
	debug("pos %2d -> (%d,%d)\n", pos, y, x)
	var edges [][]bool
	if pos <= middle {
		edges = k.Vertical
		debug("                 V")
	} else {
		edges = k.Horizontal
		debug("                 H")
	}
	if x > 0 {
		debug("(%d,%d)\n", y, x-1)
		edges[y][x-1] = true
	} else {
		debug("(%d,%d)\n", y-1, size-2)
		edges[y-1][size-2] = true
	}
}

// Read the second half of the string to set the Clue and Operation matrices.
func readSGTClues(s string, k *Puzzle) error {
	size := k.Size()
	x, y := 0, 0
	i := 0
	for i < len(s) {
		op := parseSGTOperation(s[i])
		i++
		clue := 0
		for i < len(s) {
			c := s[i]
			if c < '0' || '9' < c {
				break
			}
			d := int(c - '0')
			clue = 10*clue + d
			i++
		}
		debug("(%d,%d) %d %s\n", y, x, clue, op.Symbol())
		k.Operation[y][x] = op
		k.Clue[y][x] = clue
		x, y = nextClue(k, x, y)
		if y == size {
			break
		}
	}
	if i != len(s) {
		log.Printf("clues remaining: %s", s[i:])
		return errMalformed
	}
	return nil
}

func parseSGTOperation(c byte) Operation {
	switch c {
	case 'a':
		return Sum
	case 's':
		return Difference
	case 'm':
		return Product
	case 'd':
		return Quotient
	}
	panic(fmt.Sprintf("unexpected operation (%c)", c))
}

// Find coordinates of the next clue, which must be the
// leftmost, topmost cell in a "cage".
// It must have heavy lines on its left and top edges,
// and all cells extending to the next heavy line on the right
// must have heavy lines on their top edges.
// (All cells in column 0 implicitly have a heavy line on their left edge,
// and all cells in row 0 implicitly have a heavy line on their top edge.
// Otherwise, the Vertical and Horizontal matrices are consulted.)
func nextClue(k *Puzzle, x int, y int) (int, int) {
	size := k.Size()
	for {
		x++
		if x == size {
			x = 0
			y++
			if y == size {
				return x, y
			}
		}
		if !cageCorner(k, x, y) {
			continue
		}
		if y == 0 || x == size-1 || k.Vertical[y][x] {
			debug("yes\n")
			return x, y
		}
		if !cageTop(k, x, y) {
			continue
		}
		debug("yes\n")
		return x, y
	}
}

// cageCorner checks if a cell has heavy lines on its left and top edges.
func cageCorner(k *Puzzle, x int, y int) bool {
	debug("checking (%d,%d): ", y, x)
	if x != 0 && !k.Vertical[y][x-1] {
		debug("no line on left\n")
		return false
	}
	if y != 0 && !k.Horizontal[x][y-1] {
		debug("no line on top\n")
		return false
	}
	return true
}

// cageTop checks if all cells extending to the next heavy line on the right
// have heavy lines on their top edges.
func cageTop(k *Puzzle, x int, y int) bool {
	size := k.Size()
	for i := x + 1; i < size; i++ {
		if !k.Horizontal[i][y-1] {
			debug("no line on top of (%d,%d)\n", y, i)
			return false
		}
		debug("line on top of (%d,%d) ", y, i)
		if i == size-1 || k.Vertical[y][i] {
			return true
		}
	}
	panic("unreachable")
}

// Make a bool matrix with N rows of N-1 columns.
func makeBoolMatrix(n int) [][]bool {
	m := make([][]bool, n)
	for i := 0; i < n; i++ {
		m[i] = make([]bool, n-1)
	}
	return m
}

// Make an NxN matrix of ints.
func makeIntMatrix(n int) [][]int {
	m := make([][]int, n)
	for i := 0; i < n; i++ {
		m[i] = make([]int, n)
	}
	return m
}

// Make an NxN matrix of Operations.
func makeOpMatrix(n int) [][]Operation {
	m := make([][]Operation, n)
	for i := 0; i < n; i++ {
		m[i] = make([]Operation, n)
	}
	return m
}

// Expand a run-length encoded puzzle ID.
func decompressSGTString(s string) string {
	var b bytes.Buffer
	var prev byte
	i := 0
	repeat := 0
	for i < len(s) {
		c := s[i]
		if c == ',' {
			break
		} else if '0' <= c && c <= '9' {
			repeat = 10*repeat + int(c-'0')
		} else {
			for j := 0; j < repeat-1; j++ {
				b.WriteByte(prev)
			}
			repeat = 0
			b.WriteByte(c)
			prev = c
		}
		i++
	}
	for j := 0; j < repeat-1; j++ {
		b.WriteByte(prev)
	}
	b.WriteString(s[i:])
	return b.String()
}

func readSGTSolution(s string, k *Puzzle) error {
	f := strings.Fields(s)
	size := k.Size()
	n := size * size
	if len(f) != n {
		return fmt.Errorf("expected %d numbers in solution but got %d", n, len(f))
	}
	x, y := 0, 0
	for i := 0; i < n; i++ {
		d, err := strconv.Atoi(f[i])
		if err != nil {
			return err
		}
		k.Answer[y][x] = d
		x++
		if x == size {
			x = 0
			y++
		}
	}
	return nil
}

// Keen puzzle generator and solver programs from Simon Tatham's Portable Puzzle Collection.
const (
	genPuzzle   = "sgt-keen"
	solvePuzzle = "keensolver"
)

var (
	idFlag                 = flag.String("id", "", "play SGT Keen puzzle with specific ID")
	sizeFlag               = flag.Int("n", 6, "puzzle `size`")
	difficultyFlag         = flag.String("d", "h", "difficulty `level` (e|n|h|x|u)")
	multiplicationOnlyFlag = flag.Bool("m", false, "multiplication only")
)

func sgtPuzzle() (*Puzzle, string, error) {
	id, err := sgtPuzzleID()
	if err != nil {
		return nil, "", err
	}
	soln, err := exec.Command(solvePuzzle, id).Output()
	if err != nil {
		return nil, id, fmt.Errorf("%s: %w", solvePuzzle, err)
	}
	k, err := ReadSGT(id, string(bytes.TrimSpace(soln)))
	if err != nil {
		log.Printf("puzzle encoding: %s", id)
		log.Printf("puzzle solution:\n%s", soln)
	}
	n := *sizeFlag
	return k, fmt.Sprintf("%d × %d", n, n), err
}

func sgtPuzzleID() (string, error) {
	if *idFlag != "" {
		return *idFlag, nil
	}
	// Encode the game parameters.
	var b bytes.Buffer
	b.WriteString(strconv.Itoa(*sizeFlag))
	b.WriteRune('d')
	switch *difficultyFlag {
	case "e", "n", "h", "x", "u":
		b.WriteString(*difficultyFlag)
	default:
		return "", fmt.Errorf("difficulty (%s) must be e[asy], n[ormal], h[ard], [e]x[treme], or u[nreasonable]", *difficultyFlag)
	}
	if *multiplicationOnlyFlag {
		b.WriteRune('m')
	}
	// Generate a Keen puzzle.
	result, err := exec.Command(genPuzzle, "--generate", "1", b.String()).Output()
	if err != nil {
		return "", fmt.Errorf("%s: %w", genPuzzle, err)
	}
	return string(bytes.TrimSpace(result)), nil
}
