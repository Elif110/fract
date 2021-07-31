package oop

import (
	"fmt"

	"github.com/fract-lang/fract/pkg/obj"
)

type MapType map[Val]Val

type MapModel struct {
	M    MapType
	Defs DefMap
}

func NewMapModel() MapModel {
	var m MapModel
	m.M = MapType{}
	m.Defs.Funcs = []*Fn{
		{Name: "keys", Src: m.keysF},
		{Name: "values", Src: m.valuesF},
		{Name: "removeKey", Src: m.removeKeyF, Params: []Param{{Name: "key"}}},
	}
	return m
}

func (m *MapModel) keysF(tk obj.Token, args []*Var) Val {
	keys := NewListModel()
	for k := range m.M {
		keys.PushBack(k)
	}
	return Val{D: keys, T: List}
}

func (m *MapModel) valuesF(tk obj.Token, args []*Var) Val {
	vals := NewListModel()
	for _, v := range m.M {
		vals.PushBack(v)
	}
	return Val{D: vals, T: List}
}

func (m *MapModel) removeKeyF(tk obj.Token, args []*Var) Val {
	arg := args[0]
	_, ok := m.M[arg.V]
	delete(m.M, arg.V)
	return Val{D: fmt.Sprint(ok), T: Bool}
}
