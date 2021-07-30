package oop

import (
	"github.com/fract-lang/fract/pkg/obj"
)

type FuncCallModel interface {
	Call() *Val
	Func() *Func
}

// Class define.
type Class struct {
	F           *obj.File
	Name        string
	Constructor *Func
	Defs        DefMap
}

func (c *Class) CallConstructor(model FuncCallModel) ClassInstance {
	ci := ClassInstance{Name: c.Name, F: c.F, Defs: c.Defs}
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
	F    *obj.File
	Name string // Name of based struct.
	Defs DefMap
}
