package main

import (
	"errors"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
)

func generate(inType string) error {
	goFile := os.Getenv("GOFILE")
	if goFile == "" {
		return errors.New("$GOFILE must be set")
	}
	goPackage := os.Getenv("GOPACKAGE")
	if goPackage == "" {
		return errors.New("$GOPACKAGE must be set")
	}
	return generateFromEnv(inType, goFile, goPackage)
}

func generateFromEnv(inType string, goFile string, goPackage string) error {
	fmt.Println(inType, goFile, goPackage)
	astFile, err := parser.ParseFile(token.NewFileSet(), goFile, nil, 0)
	if err != nil {
		return err
	}
	fmt.Println(astFile.Scope)
	return nil
}

// ***** MAIN *****

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
