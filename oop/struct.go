package oop

import "github.com/fract-lang/fract/lex"

// Struct define.
type Struct struct {
	L           *lex.Lex
	Name        string
	Constructor Func
}

func (s *Struct) CallConstructor(args []Var) StructInstance {
	si := StructInstance{Name: s.Name}
	si.Fields.Vars = append(si.Fields.Vars, args...)
	si.L = s.L
	return si
}

type StructInstance struct {
	L      *lex.Lex
	Name   string // Name of based struct.
	Fields DefMap
}
