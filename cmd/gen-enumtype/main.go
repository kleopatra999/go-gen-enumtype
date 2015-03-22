package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
)

const (
	genEnumtypeAnnotiation = "@gen-enumtype"
)

var (
	ErrNil                     = errors.New("gen-enumtype: nil")
	ErrEmpty                   = errors.New("gen-enumtype: empty")
	ErrStringEmpty             = errors.New("gen-enumtype: string empty")
	ErrGofileMustBeSet         = errors.New("gen-enumtype: $GOFILE must be set")
	ErrCannotParseAnnotation   = errors.New("gen-enumtype: cannot parse annotation")
	ErrExpectedOneSpec         = errors.New("gen-enumtype: expected one spec")
	ErrExpectedTypeSpec        = errors.New("gen-enumtype: expected value spec")
	ErrExpectedStructType      = errors.New("gen-enumtype: expected struct type")
	ErrDuplicateAnnotation     = errors.New("gen-enumtype: duplicate annotation")
	ErrDuplicateAnnotationData = errors.New("gen-enumtype: duplicate annotation data")

	debug = false
)

func generate() error {
	goFile := os.Getenv("GOFILE")
	if goFile == "" {
		return ErrGofileMustBeSet
	}
	return generateFromEnv(goFile)
}

func generateFromEnv(goFile string) error {
	mode := parser.ParseComments
	if debug {
		mode |= parser.Trace
	}
	astFile, err := parser.ParseFile(token.NewFileSet(), goFile, nil, mode)
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

type annotation struct {
	enumType  string
	enumValue string
	id        uint
}

func (this annotation) empty() bool {
	return this.enumType == "" && this.enumValue == "" && this.id == 0
}

func getAnnotatedGenDecls(genDecls []*ast.GenDecl) (map[annotation]*ast.GenDecl, error) {
	annotatedGenDecls := make(map[annotation]*ast.GenDecl)
	for _, genDecl := range genDecls {
		annotation, err := getGenDeclAnnotation(genDecl)
		if err != nil {
			return nil, err
		}
		if !annotation.empty() {
			annotatedGenDecls[annotation] = genDecl
		}
	}
	return annotatedGenDecls, nil
}

func getGenDeclAnnotation(genDecl *ast.GenDecl) (annotation, error) {
	if genDecl.Doc != nil && genDecl.Doc.List != nil && len(genDecl.Doc.List) > 0 {
		for _, comment := range genDecl.Doc.List {
			if index := strings.Index(comment.Text, genEnumtypeAnnotiation); index != -1 {
				return parseAnnotation(comment.Text[index:])
			}
		}
	}
	return annotation{}, nil
}

func parseAnnotation(text string) (annotation, error) {
	split := strings.Split(text, " ")
	if len(split) != 4 {
		return annotation{}, ErrCannotParseAnnotation
	}
	id, err := strconv.Atoi(split[3])
	if err != nil {
		return annotation{}, err
	}
	return annotation{split[1], split[2], uint(id)}, nil
}

func generateFromPkgAndAnnotatedGenDecls(pkg string, annotatedGenDecls map[annotation]*ast.GenDecl) error {
	annotatedTypeSpecs, err := getAnnotatedTypeSpecs(annotatedGenDecls)
	if err != nil {
		return err
	}
	return generateFromPkgAndAnnotatedTypeSpecs(pkg, annotatedTypeSpecs)
}

func getAnnotatedTypeSpecs(annotatedGenDecls map[annotation]*ast.GenDecl) (map[annotation]*ast.TypeSpec, error) {
	annotatedTypeSpecs := make(map[annotation]*ast.TypeSpec)
	for annotation, genDecl := range annotatedGenDecls {
		if genDecl.Specs == nil {
			return nil, ErrNil
		}
		if len(genDecl.Specs) != 1 {
			return nil, ErrExpectedOneSpec
		}
		if typeSpec, ok := genDecl.Specs[0].(*ast.TypeSpec); ok {
			annotatedTypeSpecs[annotation] = typeSpec
		} else {
			return nil, ErrExpectedTypeSpec
		}
	}
	return annotatedTypeSpecs, nil
}

func generateFromPkgAndAnnotatedTypeSpecs(pkg string, annotatedTypeSpecs map[annotation]*ast.TypeSpec) error {
	annotationToStructName, err := getAnnotationToStructName(annotatedTypeSpecs)
	if err != nil {
		return err
	}
	return generateFromPkgAndAnnotationToStructName(pkg, annotationToStructName)
}

func getAnnotationToStructName(annotatedTypeSpecs map[annotation]*ast.TypeSpec) (map[annotation]string, error) {
	annotationToStructName := make(map[annotation]string)
	for annotation, typeSpec := range annotatedTypeSpecs {
		if typeSpec.Type == nil {
			return nil, ErrNil
		}
		if _, ok := typeSpec.Type.(*ast.StructType); !ok {
			return nil, ErrExpectedStructType
		}
		if typeSpec.Name == nil {
			return nil, ErrNil
		}
		if typeSpec.Name.Name == "" {
			return nil, ErrStringEmpty
		}
		if _, ok := annotationToStructName[annotation]; ok {
			return nil, ErrDuplicateAnnotation
		}
		annotationToStructName[annotation] = typeSpec.Name.Name
	}
	return annotationToStructName, nil
}

type enumValue struct {
	name       string
	id         uint
	structName string
}

func generateFromPkgAndAnnotationToStructName(pkg string, annotationToStructName map[annotation]string) error {
	enumTypeToEnumValues, err := getEnumTypeToEnumValues(annotationToStructName)
	if err != nil {
		return err
	}
	return generateFromPkgAndEnumTypeToEnumValues(pkg, enumTypeToEnumValues)
}

func getEnumTypeToEnumValues(annotationToStructName map[annotation]string) (map[string][]*enumValue, error) {
	enumTypeToEnumValues := make(map[string][]*enumValue)
	for annotation, structName := range annotationToStructName {
		if _, ok := enumTypeToEnumValues[annotation.enumType]; !ok {
			enumTypeToEnumValues[annotation.enumType] = make([]*enumValue, 0)
		}
		enumTypeToEnumValues[annotation.enumType] = append(
			enumTypeToEnumValues[annotation.enumType],
			&enumValue{
				name:       annotation.enumValue,
				id:         annotation.id,
				structName: structName,
			},
		)
	}
	if err := validateEnumTypeToEnumValues(enumTypeToEnumValues); err != nil {
		return nil, err
	}
	return enumTypeToEnumValues, nil
}

// TODO(pedge)
func validateEnumTypeToEnumValues(enumTypeToEnumValues map[string][]*enumValue) error {
	for _, enumValues := range enumTypeToEnumValues {
		seenNames := make(map[string]bool)
		seenIds := make(map[uint]bool)
		seenStructNames := make(map[string]bool)
		for _, enumValue := range enumValues {
			if _, ok := seenNames[enumValue.name]; ok {
				return ErrDuplicateAnnotationData
			}
			if _, ok := seenIds[enumValue.id]; ok {
				return ErrDuplicateAnnotationData
			}
			if _, ok := seenStructNames[enumValue.structName]; ok {
				return ErrDuplicateAnnotationData
			}
			seenNames[enumValue.name] = true
			seenIds[enumValue.id] = true
			seenStructNames[enumValue.structName] = true
		}
	}
	return nil
}

func generateFromPkgAndEnumTypeToEnumValues(pkg string, enumTypeToEnumValues map[string][]*enumValue) error {
	for enumType, enumValues := range enumTypeToEnumValues {
		fmt.Println(enumType)
		for _, enumValue := range enumValues {
			fmt.Println(enumValue)
		}
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
