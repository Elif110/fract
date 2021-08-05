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
func (m *DefMap) VarIndexByName(name string) int {
	if name[0] == '-' { // Ignore minus.
		name = name[1:]
	}
	for j, v := range m.Vars {
		if v.Name == name {
			return j
		}
	}
	return -1
}

// TYPES
// 'f' -> Function.
// 'v' -> Variable.
// DefByName returns define by name.
func (m *DefMap) DefByName(name string) (int, rune) {
	pos := m.FuncIndexByName(name)
	if pos != -1 {
		return pos, 'f'
	}
	pos = m.VarIndexByName(name)
	if pos != -1 {
		return pos, 'v'
	}
	return -1, '-'
}

// DefIndexByName returns index of name is exist name, returns -1 if not.
func (m *DefMap) DefIndexByName(name string) int {
	if name[0] == '-' { // Ignore minus.
		name = name[1:]
	}
	for _, f := range m.Funcs {
		if f.Name == name {
			return f.Line
		}
	}
	for _, v := range m.Vars {
		if v.Name == name {
			return v.Line
		}
	}
	return -1
}
