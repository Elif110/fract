package oop

import (
	"github.com/fract-lang/fract/lex"
)

// Struct define.
type Struct struct {
	L           *lex.Lex
	Name        string
	Constructor *Func
}

func (s *Struct) CallConstructor(args []*Var) StructInstance {
	si := StructInstance{Name: s.Name, L: s.L}
	si.Fields.Vars = args
	return si
}

type StructInstance struct {
	L      *lex.Lex
	Name   string // Name of based struct.
	Fields DefMap
}
