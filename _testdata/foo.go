package main

// go:generate gen-enumtype --enum_type=FooType

var (
	FooTypeOne FooType = 0
	FooTypeTwo FooType = 1
)

type FooType uint

func main() {

}
