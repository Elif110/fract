package oop

import "github.com/fract-lang/fract/pkg/obj"

// Var instance.
type Var struct {
	Name  string
	Ln    int // Line of define.
	V     Val
	Const bool
}

// Func instance.
type Func struct {
	Name          string
	Src           interface{}
	Ln            int // Line of define.
	Tks           []obj.Tokens
	Params        []Param
	DefParamCount int
}

// Param instance.
type Param struct {
	Defval Val
	Name   string
	Params bool
}
