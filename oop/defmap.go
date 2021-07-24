package oop

import "github.com/fract-lang/fract/pkg/obj"

type DefMap struct {
	Vars  []obj.Var
	Funcs []obj.Func
}

//! This code block very like to varIndexByName function.
//! If you change here, probably you must change there too.

// FuncIndexByName returns index of function by name.
func (s *DefMap) FuncIndexByName(name obj.Token) int {
	if name.V[0] == '-' { // Ignore minus.
		name.V = name.V[1:]
	}
	for j, f := range s.Funcs {
		if f.Name == name.V {
			return j
		}
	}
	return -1
}

//! This code block very like to funcIndexByName function.
//! If you change here, probably you must change there too.

// VarIndexByName returns index of variable by name.
func (s *DefMap) VarIndexByName(name obj.Token) int {
	if name.V[0] == '-' { // Ignore minus.
		name.V = name.V[1:]
	}
	for j, v := range s.Vars {
		if v.Name == name.V {
			return j
		}
	}
	return -1
}
