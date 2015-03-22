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
	enumType := flag.String("enum_type", "", "the enum type")
	flag.Parse()

	checkTrue(*enumType != "", "must specify --enum_type")
	checkError(generate(*enumType))
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
