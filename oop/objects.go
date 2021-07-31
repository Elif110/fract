package oop

import "github.com/fract-lang/fract/pkg/obj"

// Var instance.
type Var struct {
	Name string
	Line int // Line of define.
	Val  Val
}

// Fn instance.
type Fn struct {
	Name              string
	Src               interface{}
	Line              int // Line of define.
	Tokens            [][]obj.Token
	Params            []Param
	Args              []Var // Default vars.
	DefaultParamCount int
}

// Param instance.
type Param struct {
	DefaultVal Val
	Name       string
	Params     bool
}
