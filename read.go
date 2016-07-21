package kenken

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

var malformed = fmt.Errorf("malformed KenKen file")

func Read(r io.Reader) (*Puzzle, error) {
	s := bufio.NewScanner(r)

	// A section
	k := new(Puzzle)
	k.Answer = intMatrix(s, "A")
	if k.Answer == nil {
		return nil, malformed
	}

	// T section
	k.Clue = intMatrix(s, "T")
	if k.Clue == nil {
		return nil, malformed
	}

	// S section
	k.Operation = opMatrix(s, "S")
	if k.Operation == nil {
		return nil, malformed
	}

	// V section
	k.Vertical = boolMatrix(s, "V")
	if k.Vertical == nil {
		return nil, malformed
	}

	// H section
	k.Horizontal = boolMatrix(s, "H")
	if k.Horizontal == nil {
		return nil, malformed
	}

	return k, nil
}

func readMatrix(s *bufio.Scanner, label string, square bool) [][]string {
	if !s.Scan() || s.Text() != label {
		return nil
	}
	if !s.Scan() {
		return nil
	}
	line := strings.Fields(s.Text())
	nCols := len(line)
	if nCols == 0 {
		return nil
	}
	nRows := nCols
	if !square {
		nRows += 1
	}
	m := make([][]string, nRows)
	m[0] = line
	for i := 1; i < nRows; i++ {
		if !s.Scan() {
			return nil
		}
		m[i] = strings.Fields(s.Text())
		if len(m[i]) != nCols {
			return nil
		}
	}
	return m
}

func intMatrix(s *bufio.Scanner, label string) [][]int {
	m := readMatrix(s, label, true)
	if m == nil {
		return nil
	}
	a := make([][]int, len(m))
	for i, v := range m {
		a[i] = make([]int, len(v))
		for j, t := range v {
			var err error
			a[i][j], err = strconv.Atoi(t)
			if err != nil {
				return nil
			}
		}
	}
	return a
}

func boolMatrix(s *bufio.Scanner, label string) [][]bool {
	m := readMatrix(s, label, false)
	if m == nil {
		return nil
	}
	a := make([][]bool, len(m))
	for i, v := range m {
		a[i] = make([]bool, len(v))
		for j, t := range v {
			var err error
			b, err := strconv.Atoi(t)
			if err != nil {
				return nil
			}
			switch b {
			case 0:
				a[i][j] = false
			case 1:
				a[i][j] = true
			default:
				return nil
			}
		}
	}
	return a
}

func opMatrix(s *bufio.Scanner, label string) [][]Operation {
	m := readMatrix(s, label, true)
	if m == nil {
		return nil
	}
	a := make([][]Operation, len(m))
	for i, v := range m {
		a[i] = make([]Operation, len(v))
		for j, t := range v {
			switch t {
			case "0":
				a[i][j] = None
			case "1":
				a[i][j] = Given
			case "+":
				a[i][j] = Sum
			case "-":
				a[i][j] = Difference
			case "*":
				a[i][j] = Product
			case "/":
				a[i][j] = Quotient
			default:
				return nil
			}
		}
	}
	return a
}

func ReadPuzzle() (*Puzzle, string) {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s file", os.Args[0])
	}
	filename := os.Args[1]
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	k, err := Read(f)
	if err != nil {
		log.Fatal(err)
	}
	return k, filename
}
