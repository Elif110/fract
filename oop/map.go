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
		{Name: "keys", Src: m.keys},
		{Name: "vals", Src: m.vals},
		{Name: "rmkey", Src: m.rmkey, Params: []Param{{Name: "key"}}},
	}
	return m
}

func (m *MapModel) keys(tk obj.Token, args []*Var) Val {
	keys := ArrayModel{}
	for k := range m.M {
		keys = append(keys, k)
	}
	return Val{D: keys, T: Array}
}

func (m *MapModel) vals(tk obj.Token, args []*Var) Val {
	vals := ArrayModel{}
	for _, v := range m.M {
		vals = append(vals, v)
	}
	return Val{D: vals, T: Array}
}

func (m *MapModel) rmkey(tk obj.Token, args []*Var) Val {
	arg := args[0]
	_, ok := m.M[arg.V]
	delete(m.M, arg.V)
	return Val{D: fmt.Sprint(ok), T: Bool}
}
