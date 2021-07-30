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
func Exit(tk obj.Token, args []*oop.Var) {
	c := args[0].V
	if c.T != oop.Int {
		fract.Panic(tk, obj.ValuePanic, "Exit code is only be integer!")
	}
	ec, _ := strconv.ParseInt(c.String(), 10, 64)
	os.Exit(int(ec))
}

// Float convert object to float.
func Float(parameters []*oop.Var) oop.Val {
	return oop.Val{
		D: fmt.Sprintf(fract.FloatFormat, str.Conv(parameters[0].V.String())),
		T: oop.Float,
	}
}

// Input returns input from command-line.
func Input(args []*oop.Var) oop.Val {
	args[0].V.Print()
	//! Don't use fmt.Scanln
	s := bufio.NewScanner(os.Stdin)
	s.Scan()
	return oop.Val{D: s.Text(), T: oop.Str}
}

// Int convert object to integer.
func Int(args []*oop.Var) oop.Val {
	switch args[1].V.D { // Cast type.
	case "strcode":
		var v oop.ArrayModel
		for _, byt := range []byte(args[0].V.String()) {
			v = append(v, oop.Val{D: fmt.Sprint(byt), T: oop.Int})
		}
		return oop.Val{D: v, T: oop.Array}
	default: // Object.
		return oop.Val{
			D: fmt.Sprint(int(str.Conv(args[0].V.String()))),
			T: oop.Int,
		}
	}
}

// Len returns length of object.
func Len(args []*oop.Var) oop.Val {
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
	v := oop.Val{T: oop.Array}
	if szv > 0 {
		var index int
		var data oop.ArrayModel
		for ; index < szv; index++ {
			data = append(data, oop.Val{D: "0", T: oop.Int})
		}
		v.D = data
	} else {
		v.D = oop.ArrayModel{}
	}
	return v
}

// Realloc array by size.
func Realloc(tk obj.Token, args []*oop.Var) oop.Val {
	if args[0].V.T != oop.Array {
		fract.Panic(tk, obj.ValuePanic, "Value is must be array!")
	}
	szv, _ := strconv.Atoi(args[1].V.String())
	if szv < 0 {
		fract.Panic(tk, obj.ValuePanic, "Size should be minimum zero!")
	}
	var (
		data oop.ArrayModel
		b    = args[0].V.D.(oop.ArrayModel)
		v    = oop.Val{T: oop.Array}
		c    = 0
	)
	if len(b) <= szv {
		data = b
		c = len(b)
	} else {
		v.D = b[:szv]
		return v
	}
	for ; c <= szv; c++ {
		data = append(data, oop.Val{D: "0", T: oop.Int})
	}
	v.D = data
	return v
}

// Print values to cli.
func Print(tk obj.Token, args []*oop.Var) {
	for _, d := range args[0].V.D.(oop.ArrayModel) {
		if d.T == 0 {
			fract.Panic(tk, obj.ValuePanic, "Value is not printable!")
		}
		fmt.Print(d)
	}
}

// Print values to cli with new line.
func Println(tk obj.Token, args []*oop.Var) {
	Print(tk, args)
	println()
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
		return oop.Val{T: oop.Array}
	}
	t := oop.Int
	if start.T == oop.Float || to.T == oop.Float || step.T == oop.Float {
		t = oop.Float
	}
	var data oop.ArrayModel
	if startV <= toV {
		for ; startV <= toV; startV += stepV {
			data = append(data, oop.Val{D: fmt.Sprintf(fract.FloatFormat, startV), T: t})
		}
	} else {
		for ; startV >= toV; startV -= stepV {
			data = append(data, oop.Val{D: fmt.Sprintf(fract.FloatFormat, startV), T: t})
		}
	}
	return oop.Val{D: data, T: oop.Array}
}

// String convert object to string.
func String(args []*oop.Var) oop.Val {
	switch args[1].V.D {
	case "parse":
		str := ""
		if val := args[0].V; val.T == oop.Array {
			data := val.D.(oop.ArrayModel)
			if len(data) == 0 {
				str = "[]"
			} else {
				var sb strings.Builder
				sb.WriteByte('[')
				for _, data := range data {
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
		for _, d := range v.D.(oop.ArrayModel) {
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

// Append source values to destination array.
func Append(tk obj.Token, args []*oop.Var) oop.Val {
	src := args[0].V
	if src.T != oop.Array {
		fract.Panic(tk, obj.ValuePanic, "\"src\" must be array!")
	}
	src.D = append(args[0].V.D.(oop.ArrayModel), args[1].V.D.(oop.ArrayModel)...)
	return src
}

// Delete key from map.
func Del(tk obj.Token, args []*oop.Var) {
	if args[0].V.T != oop.Map {
		fract.IPanic(tk, obj.ValuePanic, `"map" parameter is must be map!`)
	}
	delete(args[0].V.D.(oop.MapModel), args[1].V)
}

func Panic(args []*oop.Var) {
	p := obj.Panic{M: args[0].V.String()}
	if fract.TryCount > 0 {
		panic(p)
	}
	fmt.Println("panic: " + p.M)
	panic("")
}

func Type(tk obj.Token, args []*oop.Var) oop.Val {
	arg := args[0]
	if arg.V.T == 0 {
		fract.Panic(tk, obj.ValuePanic, "Invalid value!")
	}
	return oop.Val{D: fmt.Sprint(arg.V.T), T: oop.Int}
}
