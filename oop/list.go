package oop

import (
	"fmt"
	"strconv"

	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
)

type TypeList []Val

type ListModel struct {
	Elems  TypeList
	Defs   DefMap
	Length int
}

func NewListModel(e ...Val) *ListModel {
	l := &ListModel{}
	l.Length = len(e)
	l.Elems = make(TypeList, l.Length)
	copy(l.Elems, e)
	l.Defs.Funcs = []*Fn{
		{Name: "pushBack", Src: l.PushBackF, Params: []Param{{Name: "v", Params: true}}},
		{Name: "pushFront", Src: l.PushFrontF, Params: []Param{{Name: "v", Params: true}}},
		{Name: "index", Src: l.IndexF, DefParamCount: 1, Params: []Param{{Name: "v"}, {Name: "start", Defval: Val{D: "0", T: Int}}}},
		{Name: "indexLast", Src: l.IndexLastF, DefParamCount: 1, Params: []Param{{Name: "v"}, {Name: "start", Defval: Val{D: "", T: Int}}}},
		{Name: "insert", Src: l.InsertF, Params: []Param{{Name: "i"}, {Name: "v", Params: true}}},
		{Name: "sub", Src: l.SubF, Params: []Param{{Name: "start"}, {Name: "to"}}},
		{Name: "removeAt", Src: l.RemoveAtF, Params: []Param{{Name: "i"}}},
		{Name: "remove", Src: l.RemoveF, DefParamCount: 1, Params: []Param{{Name: "v"}, {Name: "start", Defval: Val{D: "0", T: Int}}}},
		{Name: "removeLast", Src: l.RemoveLastF, DefParamCount: 1, Params: []Param{{Name: "v"}, {Name: "start", Defval: Val{D: "", T: Int}}}},
		{Name: "removeAll", Src: l.RemoveAllF, Params: []Param{{Name: "v"}}},
		{Name: "removeRange", Src: l.RemoveRangeF, Params: []Param{{Name: "start"}, {Name: "to"}}},
		{Name: "reverse", Src: l.ReverseF},
		{Name: "sort", Src: l.SortF, DefParamCount: 1, Params: []Param{{Name: "desc", Defval: Val{D: "false", T: Bool}}}},
		{Name: "unique", Src: l.UniqueF},
		{Name: "include", Src: l.IncludeF, DefParamCount: 1, Params: []Param{{Name: "v"}, {Name: "start", Defval: Val{D: "0", T: Int}}}},
	}
	return l
}

func (l *ListModel) PushBack(e ...Val) {
	l.Length += len(e)
	l.Elems = append(l.Elems, e...)
}

func (l *ListModel) PushBackF(tk obj.Token, args []*Var) Val {
	l.PushBack(args[0].V.D.(*ListModel).Elems...)
	return Val{}
}

func (l *ListModel) PushFrontF(tk obj.Token, args []*Var) Val {
	e := args[0].V.D.(*ListModel).Elems
	l.Length += len(e)
	l.Elems = append(e, l.Elems...)
	return Val{}
}

func (l *ListModel) IndexF(tk obj.Token, args []*Var) Val {
	iarg := args[1].V
	if iarg.T != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	i, _ := strconv.Atoi(iarg.String())
	if i < 0 || i > l.Length {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	e := args[0].V
	for ; i < l.Length; i++ {
		if l.Elems[i] == e {
			return Val{D: fmt.Sprint(i), T: Int}
		}
	}
	return Val{D: "-1", T: Int}
}

func (l *ListModel) IndexLastF(tk obj.Token, args []*Var) Val {
	iarg := args[1].V
	if iarg.T != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	var i int
	if iarg.D == "" {
		i = l.Length - 1
	} else {
		i, _ = strconv.Atoi(iarg.String())
	}
	if i < 0 || i > l.Length {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	e := args[0].V
	for ; i > 0; i-- {
		if l.Elems[i] == e {
			return Val{D: fmt.Sprint(i), T: Int}
		}
	}
	return Val{D: "-1", T: Int}
}

func (l *ListModel) IncludeF(tk obj.Token, args []*Var) Val {
	iarg := args[1].V
	if iarg.T != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	i, _ := strconv.Atoi(iarg.String())
	if i < 0 || i > l.Length {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	e := args[0].V
	for ; i < l.Length; i++ {
		if l.Elems[i] == e {
			return Val{D: "true", T: Bool}
		}
	}
	return Val{D: "false", T: Bool}
}

func (l *ListModel) InsertF(tk obj.Token, args []*Var) Val {
	iarg := args[0].V
	if iarg.T != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	i, _ := strconv.Atoi(iarg.String())
	if i < 0 || i > l.Length {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	elems := args[1].V.D.(*ListModel)
	l.Length += elems.Length
	l.Elems = append(l.Elems[:i], append(elems.Elems, l.Elems[i:]...)...)
	return Val{}
}

func (l *ListModel) SubF(tk obj.Token, args []*Var) Val {
	sarg := args[0].V
	if sarg.T != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	targ := args[1].V
	if targ.T != Int {
		fract.Panic(tk, obj.ValuePanic, "Length must be integer!")
	}
	index, _ := strconv.Atoi(sarg.String())
	if index < 0 || index > l.Length {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	len, _ := strconv.Atoi(targ.String())
	ls := NewListModel()
	if len < 0 {
		return Val{D: ls, T: List}
	} else if index+len > l.Length {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	ls.PushBack(l.Elems[index : index+len]...)
	return Val{D: ls, T: List}
}

func (l *ListModel) ReverseF(tk obj.Token, args []*Var) Val {
	for i := 0; i < l.Length/2; i++ {
		l.Elems[i], l.Elems[l.Length-i-1] = l.Elems[l.Length-i-1], l.Elems[i]
	}
	return Val{}
}

func quicksort(elems []Val) {
	// Quick Sort.
	r := len(elems)
	if r <= 1 {
		return
	}
	r--
	x := elems[r]
	i := -1
	for j := 0; j < r; j++ {
		if !elems[j].LessEquals(x) {
			continue
		}
		i++
		elems[i], elems[j] = elems[j], elems[i]
	}
	i++
	elems[i], elems[r] = elems[r], elems[i]
	quicksort(elems[:i])
	quicksort(elems[i+1:])
}

func (l *ListModel) SortF(tk obj.Token, args []*Var) Val {
	quicksort(l.Elems)
	return Val{}
}

func (l *ListModel) UniqueF(tk obj.Token, args []*Var) Val {
	ul := NewListModel()
	for _, e := range l.Elems {
		var exist bool
		for _, ue := range ul.Elems {
			if e == ue {
				exist = true
				break
			}
		}
		if !exist {
			ul.PushBack(e)
		}
	}
	return Val{D: ul, T: List}
}

func (l *ListModel) RemoveAtF(tk obj.Token, args []*Var) Val {
	iarg := args[0].V
	if iarg.T != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	i, _ := strconv.Atoi(iarg.String())
	if i < 0 || i >= l.Length {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	l.Elems = append(l.Elems[:i], l.Elems[i+1:]...)
	l.Length--
	return Val{}
}

func (l *ListModel) RemoveF(tk obj.Token, args []*Var) Val {
	iarg := args[1].V
	if iarg.T != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	i, _ := strconv.Atoi(iarg.String())
	if i < 0 || i > l.Length {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	e := args[0].V
	for ; i < l.Length; i++ {
		if l.Elems[i] == e {
			l.Elems = append(l.Elems[:i], l.Elems[i+1:]...)
			l.Length--
			return Val{D: "true", T: Bool}
		}
	}
	return Val{D: "false", T: Bool}
}

func (l *ListModel) RemoveLastF(tk obj.Token, args []*Var) Val {
	iarg := args[1].V
	if iarg.T != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	var i int
	if iarg.D == "" {
		i = l.Length - 1
	} else {
		i, _ = strconv.Atoi(iarg.String())
	}
	if i < 0 || i > l.Length {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	e := args[0].V
	for ; i > 0; i-- {
		if l.Elems[i] == e {
			l.Elems = append(l.Elems[:i], l.Elems[i+1:]...)
			l.Length--
			return Val{D: "true", T: Bool}
		}
	}
	return Val{D: "false", T: Bool}
}

func (l *ListModel) RemoveAllF(tk obj.Token, args []*Var) Val {
	e := args[0].V
	for i := 0; i < l.Length; i++ {
		if l.Elems[i] == e {
			l.Elems = append(l.Elems[:i], l.Elems[i+1:]...)
			l.Length--
			i--
		}
	}
	return Val{}
}

func (l *ListModel) RemoveRangeF(tk obj.Token, args []*Var) Val {
	sarg := args[0].V
	if sarg.T != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	targ := args[1].V
	if targ.T != Int {
		fract.Panic(tk, obj.ValuePanic, "Length must be integer!")
	}
	index, _ := strconv.Atoi(sarg.String())
	if index < 0 || index > l.Length {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	len, _ := strconv.Atoi(targ.String())
	if len < 0 {
		return Val{}
	} else if index+len > l.Length {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	l.Elems = append(l.Elems[:index], l.Elems[index+len:]...)
	l.Length -= len
	return Val{}
}
