package functions

// Built-In functions.

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fract-lang/fract/oop"
	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
	"github.com/fract-lang/fract/pkg/str"
)

// Exit from application with code.
func Exit(tk obj.Token, args []*oop.Var) oop.Val {
	code := args[0].Val
	if code.Type != oop.Int {
		fract.Panic(tk, obj.ValuePanic, "Exit code is only be integer!")
	}
	exitCode, _ := strconv.ParseInt(code.String(), 10, 64)
	os.Exit(int(exitCode))
	return oop.Val{}
}

// Float convert object to float.
func Float(tk obj.Token, args []*oop.Var) oop.Val {
	return oop.Val{
		Data: fmt.Sprintf(fract.FloatFormat, str.Conv(args[0].Val.String())),
		Type: oop.Float,
	}
}

// Input returns input from command-line.
func Input(tk obj.Token, args []*oop.Var) oop.Val {
	args[0].Val.Print()
	//! Don't use fmt.Scanln
	s := bufio.NewScanner(os.Stdin)
	s.Scan()
	return oop.Val{Data: oop.NewStringModel(s.Text()), Type: oop.String}
}

// Int convert object to integer.
func Int(tk obj.Token, args []*oop.Var) oop.Val {
	switch args[1].Val.Data { // Cast type.
	case "strcode":
		codes := oop.NewListModel()
		for _, byt := range []byte(args[0].Val.String()) {
			codes.PushBack(oop.Val{Data: fmt.Sprint(byt), Type: oop.Int})
		}
		return oop.Val{Data: codes, Type: oop.List}
	default: // Object.
		return oop.Val{
			Data: fmt.Sprint(int(str.Conv(args[0].Val.String()))),
			Type: oop.Int,
		}
	}
}

// Len returns length of object.
func Len(tk obj.Token, args []*oop.Var) oop.Val {
	return oop.Val{Data: fmt.Sprint(args[0].Val.Len()), Type: oop.Int}
}

// Calloc list by size.
func Calloc(tk obj.Token, args []*oop.Var) oop.Val {
	size := args[0].Val
	if size.Type != oop.Int {
		fract.Panic(tk, obj.ValuePanic, "Size is only be integer!")
	}
	sizeInt, _ := strconv.Atoi(size.String())
	if sizeInt < 0 {
		fract.Panic(tk, obj.ValuePanic, "Size should be minimum zero!")
	}
	value := oop.Val{Type: oop.List}
	if sizeInt > 0 {
		var index int
		list := oop.NewListModel()
		for ; index < sizeInt; index++ {
			list.PushBack(oop.Val{Data: "0", Type: oop.Int})
		}
		value.Data = list
	} else {
		value.Data = oop.NewListModel()
	}
	return value
}

// Realloc list by size.
func Realloc(tk obj.Token, args []*oop.Var) oop.Val {
	if args[0].Val.Type != oop.List {
		fract.Panic(tk, obj.ValuePanic, "Value is must be list!")
	}
	size, _ := strconv.Atoi(args[1].Val.String())
	if size < 0 {
		fract.Panic(tk, obj.ValuePanic, "Size should be minimum zero!")
	}
	var (
		list   = oop.NewListModel()
		dest   = args[0].Val.Data.(*oop.ListModel)
		val    = oop.Val{Type: oop.List}
		length = 0
	)
	if dest.Len <= size {
		list = dest
		length = dest.Len
	} else {
		val.Data = dest.Elems[:size]
		return val
	}
	for ; length <= size; length++ {
		list.PushBack(oop.Val{Data: "0", Type: oop.Int})
	}
	val.Data = list
	return val
}

// Print values to cli.
func Print(tk obj.Token, args []*oop.Var) oop.Val {
	for _, d := range args[0].Val.Data.(*oop.ListModel).Elems {
		fmt.Print(d)
	}
	return oop.Val{}
}

// Print values to cli with new line.
func Println(tk obj.Token, args []*oop.Var) oop.Val {
	Print(tk, args)
	println()
	return oop.Val{}
}

// Range returns list by parameters.
func Range(tk obj.Token, args []*oop.Var) oop.Val {
	start := args[0].Val
	to := args[1].Val
	step := args[2].Val
	if start.Type != oop.Int && start.Type != oop.Float {
		fract.Panic(tk, obj.ValuePanic, `"start" argument should be numeric!`)
	} else if to.Type != oop.Int && to.Type != oop.Float {
		fract.Panic(tk, obj.ValuePanic, `"to" argument should be numeric!`)
	} else if step.Type != oop.Int && step.Type != oop.Float {
		fract.Panic(tk, obj.ValuePanic, `"step" argument should be numeric!`)
	}
	startFloat, _ := strconv.ParseFloat(start.String(), 64)
	toFloat, _ := strconv.ParseFloat(to.String(), 64)
	stepFloat, _ := strconv.ParseFloat(step.String(), 64)
	if stepFloat <= 0 {
		return oop.Val{Type: oop.List}
	}
	typ := oop.Int
	if start.Type == oop.Float || to.Type == oop.Float || step.Type == oop.Float {
		typ = oop.Float
	}
	list := oop.NewListModel()
	if startFloat <= toFloat {
		for ; startFloat <= toFloat; startFloat += stepFloat {
			list.PushBack(oop.Val{Data: fmt.Sprintf(fract.FloatFormat, startFloat), Type: typ})
		}
	} else {
		for ; startFloat >= toFloat; startFloat -= stepFloat {
			list.PushBack(oop.Val{Data: fmt.Sprintf(fract.FloatFormat, startFloat), Type: typ})
		}
	}
	return oop.Val{Data: list, Type: oop.List}
}

// String convert object to string.
func String(tk obj.Token, args []*oop.Var) oop.Val {
	switch args[1].Val.Data {
	case "parse":
		str := ""
		if val := args[0].Val; val.Type == oop.List {
			data := val.Data.(*oop.ListModel)
			if data.Len == 0 {
				str = "[]"
			} else {
				var sb strings.Builder
				sb.WriteByte('[')
				for _, data := range data.Elems {
					sb.WriteString(data.String() + " ")
				}
				str = sb.String()[:sb.Len()-1] + "]"
			}
		} else {
			str = args[0].Val.String()
		}
		return oop.Val{Data: oop.NewStringModel(str), Type: oop.String}
	case "bytecode":
		val := args[0].Val
		var sb strings.Builder
		for _, element := range val.Data.(*oop.ListModel).Elems {
			if element.Type != oop.Int {
				sb.WriteByte(' ')
			}
			r, _ := strconv.ParseInt(element.String(), 10, 32)
			sb.WriteByte(byte(r))
		}
		return oop.Val{Data: oop.NewStringModel(sb.String()), Type: oop.String}
	default: // Object.
		arg := args[0]
		return oop.Val{Data: oop.NewStringModel(fmt.Sprintf("{data:%s type:%d}", arg.Val.Data, arg.Val.Type)), Type: oop.String}
	}
}

func Panic(tk obj.Token, args []*oop.Var) oop.Val {
	p := obj.Panic{Msg: args[0].Val.String()}
	if fract.TryCount > 0 {
		panic(p)
	}
	fmt.Println("panic: " + p.Msg)
	panic("")
}

func Type(tk obj.Token, args []*oop.Var) oop.Val {
	return oop.Val{Data: fmt.Sprint(args[0].Val.Type), Type: oop.Int}
}
