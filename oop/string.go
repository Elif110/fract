package oop

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
)

type StringModel struct {
	Value string
	Defs  DefMap
}

func NewStringModel(val string) StringModel {
	var str StringModel
	str.Value = val
	str.Defs.Funcs = []*Fn{
		{Name: "isLower", Src: str.isLowerF},
		{Name: "isUpper", Src: str.isUpperF},
		{Name: "lower", Src: str.lowerF},
		{Name: "upper", Src: str.upperF},
		{Name: "trim", Src: str.trimF},
		{Name: "trimLeft", Src: str.trimLeftF},
		{Name: "trimRight", Src: str.trimRightF},
		{Name: "sub", Src: str.subF, Params: []Param{{Name: "start"}, {Name: "len"}}},
		{Name: "index", Src: str.indexF, DefaultParamCount: 1, Params: []Param{{Name: "sub"}, {Name: "start", DefaultVal: Val{Data: "0", Type: Int}}}},
		{Name: "indexLast", Src: str.indexLastF, DefaultParamCount: 1, Params: []Param{{Name: "sub"}, {Name: "start", DefaultVal: Val{Data: "", Type: Int}}}},
		{Name: "split", Src: str.splitF, Params: []Param{{Name: "sub"}}},
		{Name: "hasPrefix", Src: str.hasPrefixF, Params: []Param{{Name: "sub"}}},
		{Name: "hasSuffix", Src: str.hasSuffixF, Params: []Param{{Name: "sub"}}},
		{Name: "replace", Src: str.replaceF, DefaultParamCount: 1, Params: []Param{{Name: "old"}, {Name: "new"}, {Name: "count", DefaultVal: Val{Data: "1", Type: Int}}}},
		{Name: "replaceAll", Src: str.replaceAllF, Params: []Param{{Name: "old"}, {Name: "new"}}},
	}
	return str
}

func (s *StringModel) isLowerF(tk obj.Token, args []Var) Val {
	for _, r := range s.Value {
		if unicode.IsLetter(r) && !unicode.IsLower(r) {
			return Val{Data: "false", Type: Bool}
		}
	}
	return Val{Data: "true", Type: Bool}
}

func (s *StringModel) isUpperF(tk obj.Token, args []Var) Val {
	for _, r := range s.Value {
		if unicode.IsLetter(r) && !unicode.IsUpper(r) {
			return Val{Data: "false", Type: Bool}
		}
	}
	return Val{Data: "true", Type: Bool}
}

func (s *StringModel) lowerF(tk obj.Token, args []Var) Val {
	return Val{Data: NewStringModel(strings.ToLower(s.Value)), Type: String}
}

func (s *StringModel) upperF(tk obj.Token, args []Var) Val {
	return Val{Data: NewStringModel(strings.ToUpper(s.Value)), Type: String}
}

func (s *StringModel) trimF(tk obj.Token, args []Var) Val {
	return Val{Data: NewStringModel(strings.TrimFunc(s.Value, unicode.IsSpace)), Type: String}
}

func (s *StringModel) trimLeftF(tk obj.Token, args []Var) Val {
	return Val{Data: NewStringModel(strings.TrimLeftFunc(s.Value, unicode.IsSpace)), Type: String}
}

func (s *StringModel) trimRightF(tk obj.Token, args []Var) Val {
	return Val{Data: NewStringModel(strings.TrimRightFunc(s.Value, unicode.IsSpace)), Type: String}
}

func (s *StringModel) subF(tk obj.Token, args []Var) Val {
	startArg := args[0].Val
	if startArg.Type != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	lenArg := args[1].Val
	if lenArg.Type != Int {
		fract.Panic(tk, obj.ValuePanic, "Length must be integer!")
	}
	index, _ := strconv.Atoi(startArg.String())
	if index < 0 || index > len(s.Value) {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	length, _ := strconv.Atoi(lenArg.String())
	if length < 0 {
		return Val{Data: NewStringModel(""), Type: String}
	} else if index+length > len(s.Value) {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	return Val{Data: NewStringModel(s.Value[index : index+length]), Type: String}
}

func (s *StringModel) indexF(tk obj.Token, args []Var) Val {
	indexArg := args[1].Val
	if indexArg.Type != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	var index int
	if indexArg.Data == "" {
		index = len(s.Value) - 1
	} else {
		index, _ = strconv.Atoi(indexArg.String())
	}
	if index < 0 || index > len(s.Value) {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	sub := args[0].Val
	if sub.Type != String {
		fract.Panic(tk, obj.OutOfRangePanic, "Value is not string!")
	}
	return Val{Data: fmt.Sprint(strings.Index(s.Value, sub.String())), Type: Int}
}

func (s *StringModel) indexLastF(tk obj.Token, args []Var) Val {
	indexArg := args[1].Val
	if indexArg.Type != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	var index int
	if indexArg.Data == "" {
		index = len(s.Value) - 1
	} else {
		index, _ = strconv.Atoi(indexArg.String())
	}
	if index < 0 || index > len(s.Value) {
		fract.Panic(tk, obj.OutOfRangePanic, "Out of range!")
	}
	sub := args[0].Val
	if sub.Type != String {
		fract.Panic(tk, obj.OutOfRangePanic, "Value is not string!")
	}
	return Val{Data: fmt.Sprint(strings.LastIndex(s.Value, sub.String())), Type: Int}
}

func (s *StringModel) splitF(tk obj.Token, args []Var) Val {
	sub := args[0].Val
	if sub.Type != String {
		fract.Panic(tk, obj.OutOfRangePanic, "Value is not string!")
	}
	list := NewListModel()
	parts := strings.Split(s.Value, sub.String())
	list.Elems = make(TypeList, len(parts))
	for i, p := range parts {
		list.Elems[i] = Val{Data: NewStringModel(p), Type: String}
	}
	return Val{Data: list, Type: List}
}

func (s *StringModel) hasPrefixF(tk obj.Token, args []Var) Val {
	sub := args[0].Val
	if sub.Type != String {
		fract.Panic(tk, obj.OutOfRangePanic, "Value is not string!")
	}
	if strings.HasPrefix(s.Value, sub.String()) {
		return Val{Data: "true", Type: Bool}
	}
	return Val{Data: "false", Type: Bool}
}

func (s *StringModel) hasSuffixF(tk obj.Token, args []Var) Val {
	sub := args[0].Val
	if sub.Type != String {
		fract.Panic(tk, obj.OutOfRangePanic, "Value is not string!")
	}
	if strings.HasSuffix(s.Value, sub.String()) {
		return Val{Data: "true", Type: Bool}
	}
	return Val{Data: "false", Type: Bool}
}

func (s *StringModel) replaceF(tk obj.Token, args []Var) Val {
	countArg := args[2].Val
	if countArg.Type != Int {
		fract.Panic(tk, obj.ValuePanic, "Start index must be integer!")
	}
	var count int
	if countArg.Data == "" {
		count = len(s.Value) - 1
	} else {
		count, _ = strconv.Atoi(countArg.String())
	}
	old := args[0].Val
	if old.Type != String {
		fract.Panic(tk, obj.OutOfRangePanic, "Value is not string!")
	}
	new := args[1].Val
	if new.Type != String {
		fract.Panic(tk, obj.OutOfRangePanic, "Value is not string!")
	}
	return Val{Data: NewStringModel(strings.Replace(s.Value, old.String(), new.String(), count)), Type: String}
}

func (s *StringModel) replaceAllF(tk obj.Token, args []Var) Val {
	old := args[0].Val
	if old.Type != String {
		fract.Panic(tk, obj.OutOfRangePanic, "Value is not string!")
	}
	new := args[1].Val
	if new.Type != String {
		fract.Panic(tk, obj.OutOfRangePanic, "Value is not string!")
	}
	return Val{Data: NewStringModel(strings.ReplaceAll(s.Value, old.String(), new.String())), Type: String}
}
