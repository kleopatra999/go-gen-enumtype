package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func generate(inType string) error {
	fmt.Println(inType)
	return nil
}

func main() {
	inType := flag.String("in_type", "", "the in type")
	flag.Parse()

	checkTrue(*inType != "", "must specify --in_type")
	checkError(generate(*inType))
	os.Exit(0)
}

func checkTrue(condition bool, message string) {
	if !condition {
		checkError(errors.New(message))
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
