package kenken

import (
	"flag"
	"log"
	"os"
)

// ReadPuzzle reads the file specified on the command line,
// or generates an SGT Keen puzzle if none is given.
func ReadPuzzle() (*Puzzle, string) {
	var k *Puzzle
	var name string
	var err error
	flag.Parse()
	args := flag.Args()
	switch len(args) {
	case 0:
		k, name, err = sgtPuzzle()
	case 1:
		name = args[0]
		var f *os.File
		f, err = os.Open(name)
		if err != nil {
			log.Fatal(err)
		}
		defer func() { _ = f.Close() }()
		k, err = Read(f)
	default:
		log.Fatalf("Usage: %s [options] [file]", os.Args[0])
	}
	if err != nil {
		log.Fatal(err)
	}
	return k, name
}
