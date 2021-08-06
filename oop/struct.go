package oop

import (
	"github.com/fract-lang/fract/lex"
	"github.com/fract-lang/fract/pkg/obj"
)

// Struct define.
type Struct struct {
	Lex         *lex.Lex
	Name        string
	Constructor *Fn
}

func (s *Struct) CallConstructor(args []VarDef) StructInstance {
	ins := StructInstance{Name: s.Name, File: s.Lex.File}
	ins.Fields.Vars = append(ins.Fields.Vars, args...)
	return ins
}

type StructInstance struct {
	File   *obj.File
	Name   string // Name of based struct.
	Fields DefMap
}
