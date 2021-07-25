package oop

type DefMap struct {
	Vars  []Var
	Funcs []Func
}

//! This code block very like to varIndexByName function.
//! If you change here, probably you must change there too.

// FuncIndexByName returns index of function by name.
func (s *DefMap) FuncIndexByName(n string) int {
	if n[0] == '-' { // Ignore minus.
		n = n[1:]
	}
	for j, f := range s.Funcs {
		if f.Name == n {
			return j
		}
	}
	return -1
}

//! This code block very like to funcIndexByName function.
//! If you change here, probably you must change there too.

// VarIndexByName returns index of variable by name.
func (s *DefMap) VarIndexByName(n string) int {
	if n[0] == '-' { // Ignore minus.
		n = n[1:]
	}
	for j, v := range s.Vars {
		if v.Name == n {
			return j
		}
	}
	return -1
}
