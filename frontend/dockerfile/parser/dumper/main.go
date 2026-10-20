package main

import (
	"fmt"
	"os"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("please supply filename(s)")
		os.Exit(1)
	}

	for _, fn := range os.Args[1:] {
		f, err := os.Open(fn) //nolint:gosec // not using os.Root to support symlinks
		if err != nil {
			panic(err)
		}
		defer f.Close()

		result, err := parser.Parse(f)
		if err != nil {
			panic(err)
		}
		fmt.Println(result.AST.Dump())
	}
}
