package oop

import (
	"github.com/fract-lang/fract/lex"
	"github.com/fract-lang/fract/pkg/obj"
)

// Struct define.
type Struct struct {
	L           *lex.Lex
	Name        string
	Constructor *Fn
}

func (s *Struct) CallConstructor(args []*Var) StructInstance {
	si := StructInstance{Name: s.Name, F: s.L.F}
	si.Fields.Vars = args
	return si
}

type StructInstance struct {
	F      *obj.File
	Name   string // Name of based struct.
	Fields DefMap
}
