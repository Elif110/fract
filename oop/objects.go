package oop

import "github.com/fract-lang/fract/pkg/obj"

// Var instance.
type Var struct {
	Name string
	Ln   int // Line of define.
	V    Val
}

// Fn instance.
type Fn struct {
	Name          string
	Src           interface{}
	Ln            int // Line of define.
	Tks           [][]obj.Token
	Params        []Param
	Args          []*Var // Default vars.
	DefParamCount int
}

// Param instance.
type Param struct {
	Defval Val
	Name   string
	Params bool
}
