package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

const (
	genEnumtypeAnnotiation = "@gen-enumtype"
)

var (
	ErrNil                   = errors.New("gen-enumtype: nil")
	ErrEmpty                 = errors.New("gen-enumtype: empty")
	ErrStringEmpty           = errors.New("gen-enumtype: string empty")
	ErrGofileMustBeSet       = errors.New("gen-enumtype: $GOFILE must be set")
	ErrCannotParseAnnotation = errors.New("gen-enumtype: cannot parse annotation")
)

func generate() error {
	goFile := os.Getenv("GOFILE")
	if goFile == "" {
		return ErrGofileMustBeSet
	}
	return generateFromEnv(goFile)
}

func generateFromEnv(goFile string) error {
	astFile, err := parser.ParseFile(token.NewFileSet(), goFile, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	return generateFromAstFile(astFile)
}

func generateFromAstFile(astFile *ast.File) error {
	pkg, err := packageFromAstFile(astFile)
	if err != nil {
		return err
	}
	genDecls, err := genDeclsFromAstFile(astFile)
	if err != nil {
		return err
	}
	return generateFromPkgAndGenDecls(pkg, genDecls)
}

func packageFromAstFile(astFile *ast.File) (string, error) {
	if astFile.Name == nil {
		return "", ErrNil
	}
	if astFile.Name.Name == "" {
		return "", ErrStringEmpty
	}
	return astFile.Name.Name, nil
}

func genDeclsFromAstFile(astFile *ast.File) ([]*ast.GenDecl, error) {
	if astFile.Decls == nil {
		return nil, ErrNil
	}
	genDecls := make([]*ast.GenDecl, 0)
	if astFile.Decls != nil {
		for _, decl := range astFile.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				genDecls = append(genDecls, genDecl)
			}
		}
	}
	return genDecls, nil
}

func generateFromPkgAndGenDecls(pkg string, genDecls []*ast.GenDecl) error {
	annotatedGenDecls, err := getAnnotatedGenDecls(genDecls)
	if err != nil {
		return err
	}
	return generateFromPkgAndAnnotatedGenDecls(pkg, annotatedGenDecls)
}

func getAnnotatedGenDecls(genDecls []*ast.GenDecl) (map[string]*ast.GenDecl, error) {
	annotatedGenDecls := make(map[string]*ast.GenDecl)
	for _, genDecl := range genDecls {
		annotation, err := getGenDeclAnnotation(genDecl)
		if err != nil {
			return nil, err
		}
		if annotation != "" {
			annotatedGenDecls[annotation] = genDecl
		}
	}
	return annotatedGenDecls, nil
}

func getGenDeclAnnotation(genDecl *ast.GenDecl) (string, error) {
	if genDecl.Doc != nil && genDecl.Doc.List != nil && len(genDecl.Doc.List) > 0 {
		for _, comment := range genDecl.Doc.List {
			if index := strings.Index(comment.Text, genEnumtypeAnnotiation); index != -1 {
				split := strings.Split(comment.Text[index:], " ")
				if len(split) != 2 {
					return "", ErrCannotParseAnnotation
				}
				return split[1], nil
			}
		}
	}
	return "", nil
}

func generateFromPkgAndAnnotatedGenDecls(pkg string, annotatedGenDecls map[string]*ast.GenDecl) error {
	for annotation, genDecl := range annotatedGenDecls {
		fmt.Println(annotation, genDecl)
	}
	return nil
}

// ***** MAIN *****

func main() {
	if err := generate(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
