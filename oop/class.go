package oop

import (
	"github.com/fract-lang/fract/lex"
)

type FuncCallModel interface {
	Call() *Val
	Func() *Func
}

// Class define.
type Class struct {
	L           *lex.Lex
	Name        string
	Constructor *Func
	Defs        DefMap
}

func (c *Class) CallConstructor(model FuncCallModel) ClassInstance {
	ci := ClassInstance{Name: c.Name, L: c.L, Defs: c.Defs}
	this := &Var{Name: "this", V: Val{D: ci, T: ClassIns, Mut: true}}
	model.Func().Args = []*Var{this}
	for _, f := range ci.Defs.Funcs {
		f.Args = append(f.Args, this)
	}
	if c.Constructor.Ln != 0 { // Call custom constructor.
		model.Call()
	}
	return ci
}

type ClassInstance struct {
	L    *lex.Lex
	Name string // Name of based struct.
	Defs DefMap
}
