package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/peter-edge/go-gen-annotatedstruct"
	"github.com/peter-edge/go-gen-common"
)

const (
	genEnumtypeAnnotation = "@gen-enumtype"
	filePrefix            = "gen_enumtype_"
)

var (
	ErrInvalidAnnotationData   = errors.New("gen-enumtype: invalid annotation data")
	ErrDuplicateAnnotationData = errors.New("gen-enumtype: duplicate annotation data")
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

func templateString() string {
	// definitions
	s := "{{$package := .Package}}"
	// package
	s += "package {{$package}}\n\n"
	// imports
	s += "import (\n"
	s += "\t\"fmt\"\n\n"
	s += "\t\"github.com/peter-edge/go-stringhelper\"\n"
	s += ")\n\n"
	s += "{{range $enumType := .EnumTypes}}"
	// type declaration
	s += "type {{$enumType.Name}}Type uint\n\n"
	// value declarations
	s += "{{range $enumValue := $enumType.EnumValues}}"
	s += "var {{$enumType.Name}}Type{{$enumValue.Name | upperCaseFirstLetter}} {{$enumType.Name}}Type = {{$enumValue.Id}}\n"
	s += "{{end}}\n"
	// type to string map
	s += "var {{$enumType.Name | lowerCaseFirstLetter}}TypeToString = map[{{$enumType.Name}}Type]string{\n"
	s += "{{range $enumValue := $enumType.EnumValues}}"
	s += "\t{{$enumType.Name}}Type{{$enumValue.Name | upperCaseFirstLetter}}: \"{{$enumValue.Name}}\",\n"
	s += "{{end}}"
	s += "}\n\n"
	// string to type map
	s += "var stringTo{{$enumType.Name}}Type = map[string]{{$enumType.Name}}Type{\n"
	s += "{{range $enumValue := $enumType.EnumValues}}"
	s += "\t\"{{$enumValue.Name}}\": {{$enumType.Name}}Type{{$enumValue.Name | upperCaseFirstLetter}},\n"
	s += "{{end}}"
	s += "}\n\n"
	// all types
	s += "func All{{$enumType.Name}}Types() []{{$enumType.Name}}Type {\n"
	s += "\treturn []{{$enumType.Name}}Type{\n"
	s += "{{range $enumValue := $enumType.EnumValues}}"
	s += "\t\t{{$enumType.Name}}Type{{$enumValue.Name | upperCaseFirstLetter}},\n"
	s += "{{end}}"
	s += "\t}\n"
	s += "}\n\n"
	// type of
	s += "func {{$enumType.Name}}TypeOf(s string) ({{$enumType.Name}}Type, error) {\n"
	s += "\t{{$enumType.Name | lowerCaseFirstLetter}}Type, ok := stringTo{{$enumType.Name}}Type[s]\n"
	s += "\tif !ok {\n"
	s += "\t\treturn 0, newErrorUnknown{{$enumType.Name}}Type(s)\n"
	s += "\t}\n"
	s += "\treturn {{$enumType.Name | lowerCaseFirstLetter}}Type, nil\n"
	s += "}\n\n"
	// string
	s += "func (this {{$enumType.Name}}Type) String() string {\n"
	s += "\tif int(this) < len({{$enumType.Name | lowerCaseFirstLetter}}TypeToString) {\n"
	s += "\t\t return {{$enumType.Name | lowerCaseFirstLetter}}TypeToString[this]\n"
	s += "\t}\n"
	s += "\tpanic(newErrorUnknown{{$enumType.Name}}Type(this).Error())\n"
	s += "}\n\n"
	// interface declaration
	s += "type {{$enumType.Name}} interface {\n"
	s += "\tfmt.Stringer\n"
	s += "\tType() {{$enumType.Name}}Type\n"
	s += "}\n\n"
	// Type() functions
	s += "{{range $enumValue := $enumType.EnumValues}}"
	s += "func (this *{{$enumValue.StructName}}) Type() {{$enumType.Name}}Type {\n"
	s += "\treturn {{$enumType.Name}}Type{{$enumValue.Name | upperCaseFirstLetter}}\n"
	s += "}\n\n"
	s += "{{end}}"
	// String() functions
	s += "{{range $enumValue := $enumType.EnumValues}}"
	s += "func (this *{{$enumValue.StructName}}) String() string {\n"
	s += "\treturn stringhelper.String(this)\n"
	s += "}\n\n"
	s += "{{end}}"
	// Switch()
	s += "func {{$enumType.Name}}Switch(\n"
	s += "\t{{$enumType.Name | lowerCaseFirstLetter}} {{$enumType.Name}},\n"
	s += "{{range $enumValue := $enumType.EnumValues}}"
	s += "\t{{$enumValue.StructName | lowerCaseFirstLetter}}Func func({{$enumValue.StructName | lowerCaseFirstLetter}} *{{$enumValue.StructName}}) error,\n"
	s += "{{end}}"
	s += ") error {\n"
	s += "\tswitch {{$enumType.Name | lowerCaseFirstLetter}}.Type() {\n"
	s += "{{range $enumValue := $enumType.EnumValues}}"
	s += "\tcase {{$enumType.Name}}Type{{$enumValue.Name | upperCaseFirstLetter}}:\n"
	s += "\t\treturn {{$enumValue.StructName | lowerCaseFirstLetter}}Func({{$enumType.Name | lowerCaseFirstLetter}}.(*{{$enumValue.StructName}}))\n"
	s += "{{end}}"
	s += "\tdefault:\n"
	s += "\t\treturn newErrorUnknown{{$enumType.Name}}Type({{$enumType.Name | lowerCaseFirstLetter}}.Type())\n"
	s += "\t}\n"
	s += "}\n\n"
	// New()
	s += "func (this {{$enumType.Name}}Type) New{{$enumType.Name}}(\n"
	s += "{{range $enumValue := $enumType.EnumValues}}"
	s += "\t{{$enumValue.StructName | lowerCaseFirstLetter}}Func func() (*{{$enumValue.StructName}}, error),\n"
	s += "{{end}}"
	s += ") ({{$enumType.Name}}, error) {\n"
	s += "\tswitch this {\n"
	s += "{{range $enumValue := $enumType.EnumValues}}"
	s += "\tcase {{$enumType.Name}}Type{{$enumValue.Name | upperCaseFirstLetter}}:\n"
	s += "\t\treturn {{$enumValue.StructName | lowerCaseFirstLetter}}Func()\n"
	s += "{{end}}"
	s += "\tdefault:\n"
	s += "\t\treturn nil, newErrorUnknown{{$enumType.Name}}Type(this)\n"
	s += "\t}\n"
	s += "}\n\n"
	// Produce()
	s += "func (this {{$enumType.Name}}Type) Produce(\n"
	s += "{{range $enumValue := $enumType.EnumValues}}"
	s += "\t{{$enumType.Name | lowerCaseFirstLetter}}Type{{$enumValue.Name | upperCaseFirstLetter}}Func func() (interface{}, error),\n"
	s += "{{end}}"
	s += ") (interface{}, error) {\n"
	s += "\tswitch this {\n"
	s += "{{range $enumValue := $enumType.EnumValues}}"
	s += "\tcase {{$enumType.Name}}Type{{$enumValue.Name | upperCaseFirstLetter}}:\n"
	s += "\t\treturn {{$enumType.Name | lowerCaseFirstLetter}}Type{{$enumValue.Name | upperCaseFirstLetter}}Func()\n"
	s += "{{end}}"
	s += "\tdefault:\n"
	s += "\t\treturn nil, newErrorUnknown{{$enumType.Name}}Type(this)\n"
	s += "\t}\n"
	s += "}\n\n"
	// Handle()
	s += "func (this {{$enumType.Name}}Type) Handle(\n"
	s += "{{range $enumValue := $enumType.EnumValues}}"
	s += "\t{{$enumType.Name | lowerCaseFirstLetter}}Type{{$enumValue.Name | upperCaseFirstLetter}}Func func() error,\n"
	s += "{{end}}"
	s += ") error {\n"
	s += "\tswitch this {\n"
	s += "{{range $enumValue := $enumType.EnumValues}}"
	s += "\tcase {{$enumType.Name}}Type{{$enumValue.Name | upperCaseFirstLetter}}:\n"
	s += "\t\treturn {{$enumType.Name | lowerCaseFirstLetter}}Type{{$enumValue.Name | upperCaseFirstLetter}}Func()\n"
	s += "{{end}}"
	s += "\tdefault:\n"
	s += "\t\treturn newErrorUnknown{{$enumType.Name}}Type(this)\n"
	s += "\t}\n"
	s += "}\n\n"
	// error
	s += "func newErrorUnknown{{$enumType.Name}}Type(value interface{}) error {\n"
	s += "\treturn fmt.Errorf(\"{{$package}}: Unknown{{$enumType.Name}}Type: %v\", value)\n"
	s += "}\n"
	s += "{{end}}"
	return s
}

func main() {
	if err := generate(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func generate() error {
	parseResult, err := annotatedstruct.ParseFromGofile(genEnumtypeAnnotation)
	if err != nil {
		return err
	}
	return generateFromParseResult(parseResult)
}

func generateFromParseResult(parseResult *annotatedstruct.ParseResult) error {
	enumTypeToEnumValues, err := getEnumTypeToEnumValues(parseResult)
	if err != nil {
		return err
	}
	return common.WriteFileFromTemplateString(filePrefix, parseResult.File, templateString(), getGenData(parseResult.Package, enumTypeToEnumValues))
}

func getEnumTypeToEnumValues(parseResult *annotatedstruct.ParseResult) (map[string][]*EnumValue, error) {
	enumTypeToEnumValues := make(map[string][]*EnumValue)
	for annotation, structDescriptor := range parseResult.AnnotationToStructDescriptor {
		split := strings.Split(annotation, " ")
		if len(split) != 4 {
			return nil, ErrInvalidAnnotationData
		}
		enumType := split[1]
		enumValueName := split[2]
		enumValueId, err := strconv.Atoi(split[3])
		if err != nil {
			return nil, err
		}
		if _, ok := enumTypeToEnumValues[enumType]; !ok {
			enumTypeToEnumValues[enumType] = make([]*EnumValue, 0)
		}
		enumTypeToEnumValues[enumType] = append(
			enumTypeToEnumValues[enumType],
			&EnumValue{
				Name:       enumValueName,
				Id:         uint(enumValueId),
				StructName: structDescriptor.Name,
			},
		)
	}
	if err := validateEnumTypeToEnumValues(enumTypeToEnumValues); err != nil {
		return nil, err
	}
	return enumTypeToEnumValues, nil
}

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

type enumValuesById []*EnumValue

func (this enumValuesById) Len() int           { return len(this) }
func (this enumValuesById) Swap(i, j int)      { this[i], this[j] = this[j], this[i] }
func (this enumValuesById) Less(i, j int) bool { return this[i].Id < this[j].Id }

func getEnumType(name string, enumValues []*EnumValue) *EnumType {
	sort.Sort(enumValuesById(enumValues))
	return &EnumType{
		Name:       name,
		EnumValues: enumValues,
	}
}
