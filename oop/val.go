package oop

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fract-lang/fract/pkg/str"
)

const (
	None      uint8 = 0
	Int       uint8 = 1
	Float     uint8 = 2
	String    uint8 = 3
	Bool      uint8 = 4
	Func      uint8 = 5
	List      uint8 = 6
	Map       uint8 = 7
	Package   uint8 = 8
	StructDef uint8 = 9
	StructIns uint8 = 10
	ClassDef  uint8 = 11
	ClassIns  uint8 = 12
)

// Val instance.
type Val struct {
	Data  interface{}
	Type  uint8
	Mut   bool
	Const bool
	Tag   string
}

// Immut returns immutable copy.
func (d Val) Immut() Val {
	val := Val{Type: d.Type}
	switch d.Type {
	case Map:
		cpy := NewMapModel()
		for k, v := range d.Data.(MapModel).Map {
			cpy.Map[*k.Get("var")] = *v.Get("var")
		}
		val.Data = cpy
	case List:
		cpy := NewListModel()
		src := *d.Data.(*ListModel)
		cpy.Elems = make(ListType, src.Len)
		for index, element := range src.Elems {
			cpy.Elems[index] = *element.Get("var")
		}
		cpy.Len = src.Len
		val.Data = cpy
	default:
		val.Data = d.Data
	}
	return val
}

// Get is returns value by mutability.
//! Val d is must be pointer!
func (v *Val) Get(valType string) *Val {
	if valType == "mut" {
		return &Val{Data: v.Data, Type: v.Type, Mut: true, Const: v.Const}
	}
	if (valType == "" && !v.Mut) || valType == "var" { //! Immutability.
		result := new(Val)
		*result = v.Immut()
		return result
	}
	return v
}

func (v Val) String() string {
	switch v.Type {
	case Func:
		return "object.func"
	case Package:
		return "object.packageref"
	case StructDef:
		return "object.struct"
	case ClassDef:
		return "object.class"
	case List:
		return fmt.Sprint(v.Data.(*ListModel).Elems)
	case Map:
		str := fmt.Sprint(v.Data.(MapModel).Map)
		return "{" + str[4:len(str)-1] + "}"
	case StructIns:
		var sb strings.Builder
		ins := v.Data.(StructInstance)
		sb.WriteString("struct{")
		for _, f := range ins.Fields.Vars {
			sb.WriteString(f.Name)
			sb.WriteRune(':')
			sb.WriteString(f.Val.String())
			sb.WriteRune(' ')
		}
		if len(ins.Fields.Vars) == 0 {
			return sb.String() + "}"
		}
		return sb.String()[:sb.Len()-1] + "}"
	case ClassIns:
		return "object.classins"
	case None:
		return "none"
	case Int, Float:
		return fmt.Sprint(v.Data)
	case Bool:
		if v.Data == true {
			return "true"
		}
		return "false"
	default:
		if v.Data == nil {
			return ""
		}
		return v.Data.(string)
	}
}

func (v Val) Print() bool {
	if v.Data == nil {
		return false
	}
	fmt.Print(v.String())
	return true
}

// IsEnum returns true value is enumerable, returns false if not.
func (v Val) IsEnum() bool {
	switch v.Type {
	case String, List, Map:
		return true
	default:
		return false
	}
}

// Length.
func (v Val) Len() int {
	switch v.Type {
	case String:
		return len(v.Data.(string))
	case List:
		return v.Data.(*ListModel).Len
	case Map:
		return len(v.Data.(MapModel).Map)
	}
	return -1
}

func (v Val) Equals(val Val) bool {
	if v.Type == val.Type {
		return v.Data == val.Data
	}
	return reflect.DeepEqual(v.Data, val.Data)
}

func (v Val) NotEquals(val Val) bool {
	return !v.Equals(val)
}

func (v Val) Greater(val Val) bool {
	return (v.Type == String && v.String() > val.String()) || (v.Type != String && str.Conv(v.String()) > str.Conv(val.String()))
}

func (v Val) Less(val Val) bool {
	return (v.Type == String && v.String() < val.String()) || (v.Type != String && str.Conv(v.String()) < str.Conv(val.String()))
}

func (v Val) GreaterEquals(val Val) bool {
	return (v.Type == String && v.String() >= val.String()) || (v.Type != String && str.Conv(v.String()) >= str.Conv(val.String()))
}

func (v Val) LessEquals(val Val) bool {
	return (v.Type == String && v.String() <= val.String()) || (v.Type != String && str.Conv(v.String()) <= str.Conv(val.String()))
}
