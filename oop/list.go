package oop

import (
	"fmt"
	"strconv"

	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
)

type TypeList []Val

type ListModel struct {
	Elems TypeList
	Defs  DefMap
	Len   int
}

func NewListModel(elems ...Val) *ListModel {
	list := &ListModel{}
	list.Len = len(elems)
	list.Elems = make(TypeList, list.Len)
	copy(list.Elems, elems)
	list.Defs.Funcs = []*Fn{
		{Name: "pushBack", Src: list.PushBackF, Params: []Param{{Name: "v", Params: true}}},
		{Name: "pushFront", Src: list.PushFrontF, Params: []Param{{Name: "v", Params: true}}},
		{Name: "index", Src: list.IndexF, DefaultParamCount: 1, Params: []Param{{Name: "v"}, {Name: "start", DefaultVal: Val{Data: "0", Type: Int}}}},
		{Name: "indexLast", Src: list.IndexLastF, DefaultParamCount: 1, Params: []Param{{Name: "v"}, {Name: "start", DefaultVal: Val{Data: "", Type: Int}}}},
		{Name: "insert", Src: list.InsertF, Params: []Param{{Name: "i"}, {Name: "v", Params: true}}},
		{Name: "sub", Src: list.SubF, Params: []Param{{Name: "start"}, {Name: "to"}}},
		{Name: "removeAt", Src: list.RemoveAtF, Params: []Param{{Name: "i"}}},
		{Name: "remove", Src: list.RemoveF, DefaultParamCount: 1, Params: []Param{{Name: "v"}, {Name: "start", DefaultVal: Val{Data: "0", Type: Int}}}},
		{Name: "removeLast", Src: list.RemoveLastF, DefaultParamCount: 1, Params: []Param{{Name: "v"}, {Name: "start", DefaultVal: Val{Data: "", Type: Int}}}},
		{Name: "removeAll", Src: list.RemoveAllF, Params: []Param{{Name: "v"}}},
		{Name: "removeRange", Src: list.RemoveRangeF, Params: []Param{{Name: "start"}, {Name: "to"}}},
		{Name: "reverse", Src: list.ReverseF},
		{Name: "sort", Src: list.SortF, DefaultParamCount: 1, Params: []Param{{Name: "desc", DefaultVal: Val{Data: "false", Type: Bool}}}},
		{Name: "unique", Src: list.UniqueF},
		{Name: "clear", Src: list.ClearF},
		{Name: "include", Src: list.IncludeF, DefaultParamCount: 1, Params: []Param{{Name: "v"}, {Name: "start", DefaultVal: Val{Data: "0", Type: Int}}}},
	}
	return list
}

func (l *ListModel) PushBack(elems ...Val) {
	l.Len += len(elems)
	l.Elems = append(l.Elems, elems...)
}

func (l *ListModel) PushBackF(tk obj.Token, args []Var) Val {
	l.PushBack(args[0].Val.Data.(*ListModel).Elems...)
	return Val{}
}

func (l *ListModel) PushFrontF(tk obj.Token, args []Var) Val {
	elems := args[0].Val.Data.(*ListModel).Elems
	l.Len += len(elems)
	l.Elems = append(elems, l.Elems...)
	return Val{}
}

func (l *ListModel) IndexF(tk obj.Token, args []Var) Val {
	indexArg := args[1].Val
	if indexArg.Type != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	index, _ := strconv.Atoi(indexArg.String())
	if index < 0 || index > l.Len {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	elem := args[0].Val
	for ; index < l.Len; index++ {
		if l.Elems[index] == elem {
			return Val{Data: fmt.Sprint(index), Type: Int}
		}
	}
	return Val{Data: "-1", Type: Int}
}

func (l *ListModel) IndexLastF(tk obj.Token, args []Var) Val {
	indexArg := args[1].Val
	if indexArg.Type != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	var index int
	if indexArg.Data == "" {
		index = l.Len - 1
	} else {
		index, _ = strconv.Atoi(indexArg.String())
	}
	if index < 0 || index > l.Len {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	elem := args[0].Val
	for ; index > 0; index-- {
		if l.Elems[index] == elem {
			return Val{Data: fmt.Sprint(index), Type: Int}
		}
	}
	return Val{Data: "-1", Type: Int}
}

func (l *ListModel) IncludeF(tk obj.Token, args []Var) Val {
	indexArg := args[1].Val
	if indexArg.Type != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	index, _ := strconv.Atoi(indexArg.String())
	if index < 0 || index > l.Len {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	elem := args[0].Val
	for ; index < l.Len; index++ {
		if l.Elems[index] == elem {
			return Val{Data: "true", Type: Bool}
		}
	}
	return Val{Data: "false", Type: Bool}
}

func (l *ListModel) InsertF(tk obj.Token, args []Var) Val {
	indexArg := args[0].Val
	if indexArg.Type != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	index, _ := strconv.Atoi(indexArg.String())
	if index < 0 || index > l.Len {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	elems := args[1].Val.Data.(*ListModel)
	l.Len += elems.Len
	l.Elems = append(l.Elems[:index], append(elems.Elems, l.Elems[index:]...)...)
	return Val{}
}

func (l *ListModel) SubF(tk obj.Token, args []Var) Val {
	startArg := args[0].Val
	if startArg.Type != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	toArg := args[1].Val
	if toArg.Type != Int {
		fract.Panic(tk, obj.ValuePanic, "Length must be integer!")
	}
	index, _ := strconv.Atoi(startArg.String())
	if index < 0 || index > l.Len {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	len, _ := strconv.Atoi(toArg.String())
	list := NewListModel()
	if len < 0 {
		return Val{Data: list, Type: List}
	} else if index+len > l.Len {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	list.PushBack(l.Elems[index : index+len]...)
	return Val{Data: list, Type: List}
}

func (l *ListModel) ReverseF(tk obj.Token, args []Var) Val {
	for i := 0; i < l.Len/2; i++ {
		l.Elems[i], l.Elems[l.Len-i-1] = l.Elems[l.Len-i-1], l.Elems[i]
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

func (l *ListModel) SortF(tk obj.Token, args []Var) Val {
	quicksort(l.Elems)
	return Val{}
}

func (l *ListModel) UniqueF(tk obj.Token, args []Var) Val {
	list := NewListModel()
	for _, elem := range l.Elems {
		var exist bool
		for _, uniqueElem := range list.Elems {
			if elem == uniqueElem {
				exist = true
				break
			}
		}
		if !exist {
			list.PushBack(elem)
		}
	}
	return Val{Data: list, Type: List}
}

func (l *ListModel) RemoveAtF(tk obj.Token, args []Var) Val {
	indexArg := args[0].Val
	if indexArg.Type != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	index, _ := strconv.Atoi(indexArg.String())
	if index < 0 || index >= l.Len {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	l.Elems = append(l.Elems[:index], l.Elems[index+1:]...)
	l.Len--
	return Val{}
}

func (l *ListModel) RemoveF(tk obj.Token, args []Var) Val {
	indexArg := args[1].Val
	if indexArg.Type != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	index, _ := strconv.Atoi(indexArg.String())
	if index < 0 || index > l.Len {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	elem := args[0].Val
	for ; index < l.Len; index++ {
		if l.Elems[index] == elem {
			l.Elems = append(l.Elems[:index], l.Elems[index+1:]...)
			l.Len--
			return Val{Data: "true", Type: Bool}
		}
	}
	return Val{Data: "false", Type: Bool}
}

func (l *ListModel) RemoveLastF(tk obj.Token, args []Var) Val {
	indexArg := args[1].Val
	if indexArg.Type != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	var index int
	if indexArg.Data == "" {
		index = l.Len - 1
	} else {
		index, _ = strconv.Atoi(indexArg.String())
	}
	if index < 0 || index > l.Len {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	elem := args[0].Val
	for ; index > 0; index-- {
		if l.Elems[index] == elem {
			l.Elems = append(l.Elems[:index], l.Elems[index+1:]...)
			l.Len--
			return Val{Data: "true", Type: Bool}
		}
	}
	return Val{Data: "false", Type: Bool}
}

func (l *ListModel) RemoveAllF(tk obj.Token, args []Var) Val {
	elem := args[0].Val
	for i := 0; i < l.Len; i++ {
		if l.Elems[i] == elem {
			l.Elems = append(l.Elems[:i], l.Elems[i+1:]...)
			l.Len--
			i--
		}
	}
	return Val{}
}

func (l *ListModel) RemoveRangeF(tk obj.Token, args []Var) Val {
	startArg := args[0].Val
	if startArg.Type != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	lenArg := args[1].Val
	if lenArg.Type != Int {
		fract.Panic(tk, obj.ValuePanic, "Length must be integer!")
	}
	index, _ := strconv.Atoi(startArg.String())
	if index < 0 || index > l.Len {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	len, _ := strconv.Atoi(lenArg.String())
	if len < 0 {
		return Val{}
	} else if index+len > l.Len {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	l.Elems = append(l.Elems[:index], l.Elems[index+len:]...)
	l.Len -= len
	return Val{}
}

func (l *ListModel) ClearF(tk obj.Token, args []Var) Val {
	l.Elems = TypeList{}
	l.Len = 0
	return Val{}
}
