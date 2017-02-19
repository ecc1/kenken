package kenken

import (
	"flag"
	"log"
	"os"
)

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
		f, err := os.Open(name)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		k, err = Read(f)
	default:
		log.Fatalf("Usage: %s [options] [file]", os.Args[0])
	}
	if err != nil {
		log.Fatal(err)
	}
	return k, name
}
