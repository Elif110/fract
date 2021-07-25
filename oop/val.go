package oop

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fract-lang/fract/pkg/str"
)

const (
	Int               uint8 = 1
	Float             uint8 = 2
	Str               uint8 = 3
	Bool              uint8 = 4
	Function          uint8 = 5
	Array             uint8 = 6
	Map               uint8 = 7
	Package           uint8 = 8
	Structure         uint8 = 9
	StructureInstance uint8 = 10
)

type ArrayModel []Val
type MapModel map[Val]Val

// Val instance.
type Val struct {
	D   interface{}
	T   uint8
	Mut bool
}

// Returns immutable copy.
func (d Val) Immut() Val {
	v := Val{T: d.T}
	switch d.T {
	case Map:
		c := MapModel{}
		for k, v := range d.D.(MapModel) {
			c[k] = v
		}
		v.D = c
	case Array:
		c := make(ArrayModel, len(d.D.(ArrayModel)))
		copy(c, d.D.(ArrayModel))
		v.D = c
	default:
		v.D = d.D
	}
	return v
}

func (d Val) String() string {
	switch d.T {
	case Function:
		return "object.func"
	case Package:
		return "object.packageref"
	case Structure:
		return "object.struct"
	case Array:
		return fmt.Sprint(d.D)
	case Map:
		s := fmt.Sprint(d.D)
		return "{" + s[4:len(s)-1] + "}"
	case StructureInstance:
		var s strings.Builder
		d := d.D.(StructInstance)
		s.WriteString("struct{")
		for _, f := range d.Fields.Vars {
			s.WriteString(f.Name)
			s.WriteRune(':')
			s.WriteString(f.V.String())
			s.WriteRune(' ')
		}
		return s.String()[:s.Len()-1] + "}"
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
	case Str, Array, Map:
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
	case Array:
		return len(v.D.(ArrayModel))
	case Map:
		return len(v.D.(MapModel))
	}
	return 0
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
