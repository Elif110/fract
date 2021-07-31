package built_in

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
	c := args[0].V
	if c.T != oop.Int {
		fract.Panic(tk, obj.ValuePanic, "Exit code is only be integer!")
	}
	ec, _ := strconv.ParseInt(c.String(), 10, 64)
	os.Exit(int(ec))
	return oop.Val{}
}

// Float convert object to float.
func Float(tk obj.Token, args []*oop.Var) oop.Val {
	return oop.Val{
		D: fmt.Sprintf(fract.FloatFormat, str.Conv(args[0].V.String())),
		T: oop.Float,
	}
}

// Input returns input from command-line.
func Input(tk obj.Token, args []*oop.Var) oop.Val {
	args[0].V.Print()
	//! Don't use fmt.Scanln
	s := bufio.NewScanner(os.Stdin)
	s.Scan()
	return oop.Val{D: s.Text(), T: oop.Str}
}

// Int convert object to integer.
func Int(tk obj.Token, args []*oop.Var) oop.Val {
	switch args[1].V.D { // Cast type.
	case "strcode":
		v := oop.NewListModel()
		for _, byt := range []byte(args[0].V.String()) {
			v.PushBack(oop.Val{D: fmt.Sprint(byt), T: oop.Int})
		}
		return oop.Val{D: v, T: oop.List}
	default: // Object.
		return oop.Val{
			D: fmt.Sprint(int(str.Conv(args[0].V.String()))),
			T: oop.Int,
		}
	}
}

// Len returns length of object.
func Len(tk obj.Token, args []*oop.Var) oop.Val {
	return oop.Val{D: fmt.Sprint(args[0].V.Len()), T: oop.Int}
}

// Calloc array by size.
func Calloc(tk obj.Token, args []*oop.Var) oop.Val {
	sz := args[0].V
	if sz.T != oop.Int {
		fract.Panic(tk, obj.ValuePanic, "Size is only be integer!")
	}
	szv, _ := strconv.Atoi(sz.String())
	if szv < 0 {
		fract.Panic(tk, obj.ValuePanic, "Size should be minimum zero!")
	}
	v := oop.Val{T: oop.List}
	if szv > 0 {
		var index int
		data := oop.NewListModel()
		for ; index < szv; index++ {
			data.PushBack(oop.Val{D: "0", T: oop.Int})
		}
		v.D = data
	} else {
		v.D = oop.NewListModel()
	}
	return v
}

// Realloc array by size.
func Realloc(tk obj.Token, args []*oop.Var) oop.Val {
	if args[0].V.T != oop.List {
		fract.Panic(tk, obj.ValuePanic, "Value is must be array!")
	}
	szv, _ := strconv.Atoi(args[1].V.String())
	if szv < 0 {
		fract.Panic(tk, obj.ValuePanic, "Size should be minimum zero!")
	}
	var (
		data = oop.NewListModel()
		b    = args[0].V.D.(*oop.ListModel)
		v    = oop.Val{T: oop.List}
		c    = 0
	)
	if b.Length <= szv {
		data = b
		c = b.Length
	} else {
		v.D = b.Elems[:szv]
		return v
	}
	for ; c <= szv; c++ {
		data.PushBack(oop.Val{D: "0", T: oop.Int})
	}
	v.D = data
	return v
}

// Print values to cli.
func Print(tk obj.Token, args []*oop.Var) oop.Val {
	for _, d := range args[0].V.D.(*oop.ListModel).Elems {
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

// Range returns array by parameters.
func Range(tk obj.Token, args []*oop.Var) oop.Val {
	start := args[0].V
	to := args[1].V
	step := args[2].V
	if start.T != oop.Int && start.T != oop.Float {
		fract.Panic(tk, obj.ValuePanic, `"start" argument should be numeric!`)
	} else if to.T != oop.Int && to.T != oop.Float {
		fract.Panic(tk, obj.ValuePanic, `"to" argument should be numeric!`)
	} else if step.T != oop.Int && step.T != oop.Float {
		fract.Panic(tk, obj.ValuePanic, `"step" argument should be numeric!`)
	}
	if start.T != oop.Int && start.T != oop.Float || to.T != oop.Int &&
		to.T != oop.Float || step.T != oop.Int && step.T != oop.Float {
		fract.Panic(tk, obj.ValuePanic, "Values should be integer or float!")
	}
	startV, _ := strconv.ParseFloat(start.String(), 64)
	toV, _ := strconv.ParseFloat(to.String(), 64)
	stepV, _ := strconv.ParseFloat(step.String(), 64)
	if stepV <= 0 {
		return oop.Val{T: oop.List}
	}
	t := oop.Int
	if start.T == oop.Float || to.T == oop.Float || step.T == oop.Float {
		t = oop.Float
	}
	data := oop.NewListModel()
	if startV <= toV {
		for ; startV <= toV; startV += stepV {
			data.PushBack(oop.Val{D: fmt.Sprintf(fract.FloatFormat, startV), T: t})
		}
	} else {
		for ; startV >= toV; startV -= stepV {
			data.PushBack(oop.Val{D: fmt.Sprintf(fract.FloatFormat, startV), T: t})
		}
	}
	return oop.Val{D: data, T: oop.List}
}

// String convert object to string.
func String(tk obj.Token, args []*oop.Var) oop.Val {
	switch args[1].V.D {
	case "parse":
		str := ""
		if val := args[0].V; val.T == oop.List {
			data := val.D.(*oop.ListModel)
			if data.Length == 0 {
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
			str = args[0].V.String()
		}
		return oop.Val{D: str, T: oop.Str}
	case "bytecode":
		v := args[0].V
		var sb strings.Builder
		for _, d := range v.D.(*oop.ListModel).Elems {
			if d.T != oop.Int {
				sb.WriteByte(' ')
			}
			r, _ := strconv.ParseInt(d.String(), 10, 32)
			sb.WriteByte(byte(r))
		}
		return oop.Val{D: sb.String(), T: oop.Str}
	default: // Object.
		arg := args[0]
		return oop.Val{D: fmt.Sprintf("{data:%s type:%d}", arg.V.D, arg.V.T), T: oop.Str}
	}
}

func Panic(tk obj.Token, args []*oop.Var) oop.Val {
	p := obj.Panic{M: args[0].V.String()}
	if fract.TryCount > 0 {
		panic(p)
	}
	fmt.Println("panic: " + p.M)
	panic("")
}

func Type(tk obj.Token, args []*oop.Var) oop.Val {
	return oop.Val{D: fmt.Sprint(args[0].V.T), T: oop.Int}
}
