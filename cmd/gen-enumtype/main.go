package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"os"
	"strconv"
	"strings"
	"unicode"
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
	ErrFileDoesNotEndInDotGo   = errors.New("gen-enumtype: file does not end in .go")

	debug = false
)

type GenData struct {
	Package   string
	EnumTypes []*EnumType
}

type EnumType struct {
	Name       string
	EnumValues []*EnumValue
}

type EnumValue struct {
	Name       string
	Id         uint
	StructName string
}

var templateString = `
package {{.Package}}

{{range $enumType := .EnumTypes}}type {{$enumType.Name}}Type uint

{{range $enumValue := $enumType.EnumValues}}var {{$enumType.Name}}Type{{$enumValue.Name | upperCaseFirstLetter}} {{$enumType.Name}} = {{$enumValue.Id}}
{{end}}
func All{{$enumType.Name}}Types() []{{$enumType.Name}}Type {
	return []{{$enumType.Name}}Type{
{{range $enumValue := $enumType.EnumValues}}		{{$enumType.Name}}Type{{$enumValue.Name | upperCaseFirstLetter}},
{{end}}
	}
}
{{end}}
`

func main() {
	if err := generate(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

type annotation struct {
	enumType  string
	enumValue string
	id        uint
}

func (this annotation) empty() bool {
	return this.enumType == "" && this.enumValue == "" && this.id == 0
}

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
	return generateFromAstFile(goFile, astFile)
}

func generateFromAstFile(goFile string, astFile *ast.File) error {
	pkg, err := packageFromAstFile(astFile)
	if err != nil {
		return err
	}
	genDecls, err := genDeclsFromAstFile(astFile)
	if err != nil {
		return err
	}
	return generateFromGenDecls(goFile, pkg, genDecls)
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

func generateFromGenDecls(goFile string, pkg string, genDecls []*ast.GenDecl) error {
	annotatedGenDecls, err := getAnnotatedGenDecls(genDecls)
	if err != nil {
		return err
	}
	return generateFromAnnotatedGenDecls(goFile, pkg, annotatedGenDecls)
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

func generateFromAnnotatedGenDecls(goFile string, pkg string, annotatedGenDecls map[annotation]*ast.GenDecl) error {
	annotatedTypeSpecs, err := getAnnotatedTypeSpecs(annotatedGenDecls)
	if err != nil {
		return err
	}
	return generateFromAnnotatedTypeSpecs(goFile, pkg, annotatedTypeSpecs)
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

func generateFromAnnotatedTypeSpecs(goFile string, pkg string, annotatedTypeSpecs map[annotation]*ast.TypeSpec) error {
	annotationToStructName, err := getAnnotationToStructName(annotatedTypeSpecs)
	if err != nil {
		return err
	}
	return generateFromAnnotationToStructName(goFile, pkg, annotationToStructName)
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

func generateFromAnnotationToStructName(goFile string, pkg string, annotationToStructName map[annotation]string) error {
	enumTypeToEnumValues, err := getEnumTypeToEnumValues(annotationToStructName)
	if err != nil {
		return err
	}
	return generateFromEnumTypeToEnumValues(goFile, pkg, enumTypeToEnumValues)
}

func getEnumTypeToEnumValues(annotationToStructName map[annotation]string) (map[string][]*EnumValue, error) {
	enumTypeToEnumValues := make(map[string][]*EnumValue)
	for annotation, structName := range annotationToStructName {
		if _, ok := enumTypeToEnumValues[annotation.enumType]; !ok {
			enumTypeToEnumValues[annotation.enumType] = make([]*EnumValue, 0)
		}
		enumTypeToEnumValues[annotation.enumType] = append(
			enumTypeToEnumValues[annotation.enumType],
			&EnumValue{
				Name:       annotation.enumValue,
				Id:         annotation.id,
				StructName: structName,
			},
		)
	}
	if err := validateEnumTypeToEnumValues(enumTypeToEnumValues); err != nil {
		return nil, err
	}
	return enumTypeToEnumValues, nil
}

// TODO(pedge)
func validateEnumTypeToEnumValues(enumTypeToEnumValues map[string][]*EnumValue) error {
	for _, enumValues := range enumTypeToEnumValues {
		seenNames := make(map[string]bool)
		seenIds := make(map[uint]bool)
		seenStructNames := make(map[string]bool)
		for _, enumValue := range enumValues {
			if _, ok := seenNames[enumValue.Name]; ok {
				return ErrDuplicateAnnotationData
			}
			if _, ok := seenIds[enumValue.Id]; ok {
				return ErrDuplicateAnnotationData
			}
			if _, ok := seenStructNames[enumValue.StructName]; ok {
				return ErrDuplicateAnnotationData
			}
			seenNames[enumValue.Name] = true
			seenIds[enumValue.Id] = true
			seenStructNames[enumValue.StructName] = true
		}
	}
	return nil
}

func generateFromEnumTypeToEnumValues(goFile string, pkg string, enumTypeToEnumValues map[string][]*EnumValue) error {
	return generateFromGenData(goFile, getGenData(pkg, enumTypeToEnumValues))
}

func getGenData(pkg string, enumTypeToEnumValues map[string][]*EnumValue) *GenData {
	enumTypes := make([]*EnumType, len(enumTypeToEnumValues))
	i := 0
	for enumTypeName, enumValues := range enumTypeToEnumValues {
		enumTypes[i] = getEnumType(enumTypeName, enumValues)
		i++
	}
	return &GenData{
		Package:   pkg,
		EnumTypes: enumTypes,
	}
}

func getEnumType(name string, enumValues []*EnumValue) *EnumType {
	return &EnumType{
		Name:       name,
		EnumValues: enumValues,
	}
}

func generateFromGenData(goFile string, genData *GenData) (retErr error) {
	if !strings.HasSuffix(goFile, ".go") {
		// lol
		return ErrFileDoesNotEndInDotGo
	}
	outputFile := goFile[0:(len(goFile)-3)] + "_gen_enumtype.go"
	output, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer func() {
		if err := output.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	template := template.New("root")
	template.Funcs(
		map[string]interface{}{
			"lowerCaseFirstLetter": lowerCaseFirstLetter,
			"upperCaseFirstLetter": upperCaseFirstLetter,
		},
	)
	if _, err := template.Parse(strings.TrimSpace(templateString)); err != nil {
		return err
	}
	return template.Execute(output, genData)
}

func lowerCaseFirstLetter(s string) string {
	a := []rune(s)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}

func upperCaseFirstLetter(s string) string {
	a := []rune(s)
	a[0] = unicode.ToUpper(a[0])
	return string(a)
}
