package jxl

type Visitor interface {
	Visit(fields []Fields) error
	Bool(defaultValue bool, value *bool) error
	U32() error //encoder?
}
