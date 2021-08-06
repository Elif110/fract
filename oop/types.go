package oop

import "github.com/fract-lang/fract/pkg/obj"

type ListType []Val
type MapType map[Val]Val
type VarDef *Var
type BuiltInFuncType func(obj.Token, []VarDef) Val
