package kenken

import (
	"flag"
	"log"
	"os"
	"path"
)

// ReadPuzzle reads the file specified on the command line,
// or generates an SGT Keen puzzle if none is given.
func ReadPuzzle() (*Puzzle, string) {
	var k *Puzzle
	var title string
	var err error
	flag.Parse()
	args := flag.Args()
	switch len(args) {
	case 0:
		k, title, err = sgtPuzzle()
	case 1:
		filename := args[0]
		var f *os.File
		f, err = os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		title = path.Base(filename)
		k, err = Read(f)
	default:
		log.Fatalf("Usage: %s [options] [file]", os.Args[0])
	}
	if err != nil {
		log.Fatal(err)
	}
	return k, title
}
