package oop

import (
	"fmt"

	"github.com/fract-lang/fract/pkg/obj"
)

type MapModel struct {
	Map  MapType
	Defs DefMap
}

func NewMapModel() MapModel {
	var m MapModel
	m.Map = MapType{}
	m.Defs.Funcs = []*Fn{
		{Name: "keys", Src: m.keysF},
		{Name: "values", Src: m.valuesF},
		{Name: "removeKey", Src: m.removeKeyF, Params: []Param{{Name: "key"}}},
	}
	return m
}

func (m *MapModel) keysF(tk obj.Token, args []VarDef) Val {
	keys := NewListModel()
	for key := range m.Map {
		keys.PushBack(key)
	}
	return Val{Data: keys, Type: List}
}

func (m *MapModel) valuesF(tk obj.Token, args []VarDef) Val {
	vals := NewListModel()
	for _, val := range m.Map {
		vals.PushBack(val)
	}
	return Val{Data: vals, Type: List}
}

func (m *MapModel) removeKeyF(tk obj.Token, args []VarDef) Val {
	keyArg := args[0]
	_, ok := m.Map[keyArg.Val]
	delete(m.Map, keyArg.Val)
	return Val{Data: fmt.Sprint(ok), Type: Bool}
}
