package kenken

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var malformed = fmt.Errorf("malformed KenKen file")

func Read(r io.Reader) (*Puzzle, error) {
	s := bufio.NewScanner(r)
	k := new(Puzzle)
	k.Answer = intMatrix(s, "A")
	k.Clue = intMatrix(s, "T")
	k.Operation = opMatrix(s, "S")
	k.Vertical = boolMatrix(s, "V")
	k.Horizontal = boolMatrix(s, "H")
	if k.Answer == nil || k.Clue == nil || k.Operation == nil || k.Vertical == nil || k.Horizontal == nil {
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
