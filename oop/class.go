package oop

import (
	"github.com/fract-lang/fract/pkg/obj"
)

type FuncCallModel interface {
	Call() *Val
	Func() *Fn
}

// Class define.
type Class struct {
	File        *obj.File
	Name        string
	Constructor *Fn
	Defs        DefMap
}

func (c *Class) CallConstructor(model FuncCallModel) ClassInstance {
	ins := ClassInstance{Name: c.Name, File: c.File, Defs: c.Defs}
	this := &Var{Name: "this", Val: Val{Data: ins, Type: ClassIns, Mut: true}}
	model.Func().Args = []VarDef{this}
	for _, fn := range ins.Defs.Funcs {
		fn.Args = append(fn.Args, this)
	}
	if c.Constructor.Line != 0 { // Call custom constructor.
		model.Call()
	}
	return ins
}

type ClassInstance struct {
	File *obj.File
	Name string // Name of based struct.
	Defs DefMap
}
