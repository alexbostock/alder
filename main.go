package main

import (
	"bufio"
	"io/ioutil"
	"os"

	"github.com/alexbostock/alder/database"
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
	db := database.New(4, schema)

	r := bufio.NewReader(os.Stdin)
	for {
		line, err := r.ReadString(';')
		if err != nil {
			if err.Error() == "EOF" {
				os.Exit(0)
			} else {
				panic(err)
			}
		}

		db.Query(line[:len(line)-1])
	}

	_ = schema
}
