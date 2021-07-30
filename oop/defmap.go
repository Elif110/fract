package oop

type DefMap struct {
	Vars  []*Var
	Funcs []*Fn
}

// FuncIndexByName returns index of function by name.
func (m *DefMap) FuncIndexByName(n string) int {
	if n[0] == '-' { // Ignore minus.
		n = n[1:]
	}
	for j, f := range m.Funcs {
		if f.Name == n {
			return j
		}
	}
	return -1
}

// VarIndexByName returns index of variable by name.
func (m *DefMap) VarIndexByName(n string) int {
	if n[0] == '-' { // Ignore minus.
		n = n[1:]
	}
	for j, v := range m.Vars {
		if v.Name == n {
			return j
		}
	}
	return -1
}

// TYPES
// 'f' -> Function.
// 'v' -> Variable.
// DefByName returns define by name.
func (m *DefMap) DefByName(n string) (int, rune) {
	pos := m.FuncIndexByName(n)
	if pos != -1 {
		return pos, 'f'
	}
	pos = m.VarIndexByName(n)
	if pos != -1 {
		return pos, 'v'
	}
	return -1, '-'
}

// DefinedName returns index of name is exist name, returns -1 if not.
func (m *DefMap) DefinedName(n string) int {
	if n[0] == '-' { // Ignore minus.
		n = n[1:]
	}
	for _, f := range m.Funcs {
		if f.Name == n {
			return f.Ln
		}
	}
	for _, v := range m.Vars {
		if v.Name == n {
			return v.Ln
		}
	}
	return -1
}
