package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/alexbostock/alder/schema"
	"github.com/alexbostock/alder/sql"
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

		sql.Compile(schema, line[:len(line)-1])
		fmt.Println("(Query processor is not yet implemented)")
	}

	_ = schema
}
