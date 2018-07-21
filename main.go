package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	//"github.com/davecgh/go-spew/spew"
	"github.com/jxskiss/gothrifter/generator"
	"github.com/jxskiss/gothrifter/parser"
)

func main() {
	flags := flag.NewFlagSet("thriferc", flag.ExitOnError)
	filename := flags.String("idl", "", "the thrift IDL file")
	prefix := flags.String("prefix", "gen-thrifter", "the prefix of import path")
	output := flags.String("output", "gen-thrifter", "the root output path for generated files")
	genAll := flags.Bool("all", false, "also generate all included thrift files")
	flags.Parse(os.Args[1:])

	if filepath.Base(*prefix) != filepath.Base(*output) {
		*prefix = filepath.Join(*prefix, filepath.Base(*output))
	}
	*output = parser.AbsPath(*output)
	if *filename == "" {
		fmt.Fprintln(os.Stderr, "ERROR: no idl file specified")
		flags.PrintDefaults()
		os.Exit(1)
	}

	*filename = parser.AbsPath(*filename)
	g := generator.New(*filename, *prefix, *output)
	if *genAll {
		g.GenAll = true
	}
	err := g.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse:", err)
		os.Exit(1)
	}

	err = g.Generate()
	if err != nil {
		fmt.Fprintln(os.Stderr, "generate:", err)
		os.Exit(1)
	}

	//log.Println(spew.Sdump(g.RootPkg))
}
