package main

import (
	"io/ioutil"
	"os"

	"github.com/alexbostock/alder/schema"
)

func main() {
	if len(os.Args) < 2 {
		os.Stderr.WriteString("usage: alder schemaFileName\n")
		os.Exit(1)
	}

	schemaFileName := os.Args[1]
	schemaFile, err := ioutil.ReadFile(schemaFileName)
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(2)
	}

	schema := schema.New(schemaFile)

	_ = schema
}
