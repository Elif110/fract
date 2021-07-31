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
	Str       uint8 = 3
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
	D     interface{}
	T     uint8
	Mut   bool
	Const bool
}

// Returns immutable copy.
func (d Val) Immut() Val {
	v := Val{T: d.T}
	switch d.T {
	case Map:
		c := NewMapModel()
		for k, v := range d.D.(MapModel).M {
			c.M[k] = v
		}
		v.D = c
	case List:
		c := NewListModel()
		src := *d.D.(*ListModel)
		c.Elems = make(TypeList, src.Length)
		copy(c.Elems, src.Elems)
		c.Length = src.Length
		v.D = c
	default:
		v.D = d.D
	}
	return v
}

func (d Val) String() string {
	switch d.T {
	case Func:
		return "object.func"
	case Package:
		return "object.packageref"
	case StructDef:
		return "object.struct"
	case ClassDef:
		return "object.class"
	case List:
		return fmt.Sprint(d.D.(*ListModel).Elems)
	case Map:
		s := fmt.Sprint(d.D.(MapModel).M)
		return "{" + s[4:len(s)-1] + "}"
	case StructIns:
		var s strings.Builder
		d := d.D.(StructInstance)
		s.WriteString("struct{")
		for _, f := range d.Fields.Vars {
			s.WriteString(f.Name)
			s.WriteRune(':')
			s.WriteString(f.V.String())
			s.WriteRune(' ')
		}
		if len(d.Fields.Vars) == 0 {
			return s.String() + "}"
		}
		return s.String()[:s.Len()-1] + "}"
	case ClassIns:
		return "object.classins"
	case None:
		return "none"
	default:
		if d.D == nil {
			return ""
		}
		return d.D.(string)
	}
}

func (v Val) Print() bool {
	if v.D == nil {
		return false
	}
	fmt.Print(v.String())
	return true
}

// Is enumerable?
func (v Val) IsEnum() bool {
	switch v.T {
	case Str, List, Map:
		return true
	default:
		return false
	}
}

// Length.
func (v Val) Len() int {
	switch v.T {
	case Str:
		return len(v.D.(string))
	case List:
		return v.D.(*ListModel).Length
	case Map:
		return len(v.D.(MapModel).M)
	}
	return -1
}

func (v Val) Equals(dt Val) bool {
	return reflect.DeepEqual(v.D, dt.D)
}

func (v Val) NotEquals(dt Val) bool {
	return !v.Equals(dt)
}

func (v Val) Greater(dt Val) bool {
	return (v.T == Str && v.String() > dt.String()) || (v.T != Str && str.Conv(v.String()) > str.Conv(dt.String()))
}

func (v Val) Less(dt Val) bool {
	return (v.T == Str && v.String() < dt.String()) || (v.T != Str && str.Conv(v.String()) < str.Conv(dt.String()))
}

func (v Val) GreaterEquals(dt Val) bool {
	return (v.T == Str && v.String() >= dt.String()) || (v.T != Str && str.Conv(v.String()) >= str.Conv(dt.String()))
}

func (v Val) LessEquals(dt Val) bool {
	return (v.T == Str && v.String() <= dt.String()) || (v.T != Str && str.Conv(v.String()) <= str.Conv(dt.String()))
}
