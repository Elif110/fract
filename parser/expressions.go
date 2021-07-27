package parser

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/fract-lang/fract/oop"
	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
	"github.com/fract-lang/fract/pkg/str"
)

// Compare arithmetic values.
func compVals(opr string, d0, d1 oop.Val) bool {
	if d0.T != d1.T && (d0.T == oop.Str || d1.T == oop.Str) {
		return false
	}
	switch opr {
	case "==": // Equals.
		if !d0.Equals(d1) {
			return false
		}
	case "<>": // Not equals.
		if !d0.NotEquals(d1) {
			return false
		}
	case ">": // Greater.
		if !d0.Greater(d1) {
			return false
		}
	case "<": // Less.
		if !d0.Less(d1) {
			return false
		}
	case ">=": // Greater or equals.
		if !d0.GreaterEquals(d1) {
			return false
		}
	case "<=": // Less or equals.
		if !d0.LessEquals(d1) {
			return false
		}
	}
	return true
}

// Compare values.
func comp(v0, v1 oop.Val, opr obj.Token) bool {
	// In.
	if opr.V == "in" {
		if v1.T != oop.Array && v1.T != oop.Str && v1.T != oop.Map {
			fract.IPanic(opr, obj.ValuePanic, "Value is can should be string, array or map!")
		}
		switch v1.T {
		case oop.Array:
			dt := v0.String()
			for _, d := range v1.D.(oop.ArrayModel) {
				if strings.Contains(d.String(), dt) {
					return true
				}
			}
			return false
		case oop.Map:
			_, ok := v1.D.(oop.MapModel)[v0]
			return ok
		}
		// String.
		if v0.T == oop.Array {
			dt := v1.String()
			for _, d := range v0.D.(oop.ArrayModel) {
				if d.T != oop.Str {
					fract.IPanic(opr, obj.ValuePanic, "All values is not string!")
				}
				if strings.Contains(dt, d.String()) {
					return true
				}
			}
		} else {
			if v1.T != oop.Str {
				fract.IPanic(opr, obj.ValuePanic, "All datas is not string!")
			}
			if strings.Contains(v1.String(), v0.String()) {
				return true
			}
		}
		return false
	}
	return compVals(opr.V, v0, v1)
}

// procCondition returns condition result.
func (p *Parser) procCondition(tks []obj.Token) string {
	T := oop.Val{D: "true", T: oop.Bool}
	// Process condition.
	ors := conditionalProcesses(tks, "||")
	for _, or := range ors {
		// Decompose and conditions.
		ands := conditionalProcesses(or, "&&")
		// Is and long statement?
		if len(ands) > 1 {
			for _, and := range ands {
				i, opr := findConditionOpr(and)
				// Operator is not found?
				if i == -1 {
					opr.V = "=="
					if comp(*p.procValTks(and), T, opr) {
						return "true"
					}
					return "false"
				}
				// Operator is first or last?
				if i == 0 {
					fract.IPanic(and[0], obj.SyntaxPanic, "Comparison values are missing!")
				} else if i == len(and)-1 {
					fract.IPanic(and[len(and)-1], obj.SyntaxPanic, "Comparison values are missing!")
				}
				if !comp(*p.procValTks(and[:i]), *p.procValTks(and[i+1:]), opr) {
					return "false"
				}
			}
			return "true"
		}
		i, opr := findConditionOpr(or)
		// Operator is not found?
		if i == -1 {
			opr.V = "=="
			if comp(*p.procValTks(or), T, opr) {
				return "true"
			}
			continue
		}
		// Operator is first or last?
		if i == 0 {
			fract.IPanic(or[0], obj.SyntaxPanic, "Comparison values are missing!")
		} else if i == len(or)-1 {
			fract.IPanic(or[len(or)-1], obj.SyntaxPanic, "Comparison values are missing!")
		}
		if comp(*p.procValTks(or[:i]), *p.procValTks(or[i+1:]), opr) {
			return "true"
		}
	}
	return "false"
}

// Get string arithmetic compatible data.
func arith(tks obj.Token, d oop.Val) string {
	ret := d.String()
	switch d.T {
	case oop.Function,
		oop.Package,
		oop.StructDef,
		oop.ClassDef,
		oop.ClassIns:
		fract.IPanic(tks, obj.ArithmeticPanic, "\""+ret+"\" is not compatible with arithmetic processes!")
	case oop.Map:
		fract.IPanic(tks, obj.ArithmeticPanic, "\"object.map\" is not compatible with arithmetic processes!")
	case oop.StructIns:
		fract.IPanic(tks, obj.ArithmeticPanic, "\"object.structins\" is not compatible with arithmetic processes!")
	}
	return ret
}

// process instance for solver.
type process struct {
	f   []obj.Token // Tokens of first oop.
	fv  oop.Val     // Value instance of first oop.
	s   []obj.Token // Tokens of second oop.
	sv  oop.Val     // Value instance of second oop.
	opr obj.Token   // Operator of process.
}

// solve process.
func solve(opr obj.Token, a, b float64) float64 {
	var r float64
	switch opr.V {
	case "+": // Addition.
		r = a + b
	case "-": // Subtraction.
		r = a - b
	case "*": // Multiply.
		r = a * b
	case "/", "//": // Division.
		if a == 0 || b == 0 {
			fract.Panic(opr, obj.DivideByZeroPanic, "Divide by zero!")
		}
		r = a / b
	case "|": // Binary or.
		r = float64(int(a) | int(b))
	case "&": // Binary and.
		r = float64(int(a) & int(b))
	case "^": // Bitwise exclusive or.
		r = float64(int(a) ^ int(b))
	case "**": // Exponentiation.
		r = math.Pow(a, b)
	case "%": // Mod.
		r = math.Mod(a, b)
	case "<<": // Left shift.
		if b < 0 {
			fract.IPanic(opr, obj.ArithmeticPanic, "Shifter is cannot should be negative!")
		}
		r = float64(int(a) << int(b))
	case ">>": // Right shift.
		if b < 0 {
			fract.IPanic(opr, obj.ArithmeticPanic, "Shifter is cannot should be negative!")
		}
		r = float64(int(a) >> int(b))
	default:
		fract.IPanic(opr, obj.SyntaxPanic, "Operator is invalid!")
	}
	return r
}

// Check data and set ready.
func readyData(p process, d oop.Val) oop.Val {
	if p.fv.T == oop.Str || p.sv.T == oop.Str {
		d.T = oop.Str
	} else if p.opr.V == "/" || p.fv.T == oop.Float || p.sv.T == oop.Float {
		d.T = oop.Float
		return d
	}
	return d
}

// solveProc solve arithmetic process.
func solveProc(p process) oop.Val {
	v := oop.Val{D: "0", T: oop.Int}
	fl := p.fv.Len()
	sl := p.sv.Len()
	// String?
	if (fl != 0 && p.fv.T == oop.Str) || (sl != 0 && p.sv.T == oop.Str) {
		if p.fv.T == p.sv.T { // Both string?
			v.T = oop.Str
			switch p.opr.V {
			case "+":
				v.D = p.fv.String() + p.sv.String()
			case "-":
				flen := len(p.fv.String())
				slen := len(p.sv.String())
				if flen == 0 || slen == 0 {
					v.D = ""
					break
				}
				if flen == 1 && slen > 1 {
					r, _ := strconv.ParseInt(p.fv.String(), 10, 32)
					fr := rune(r)
					for _, r := range p.sv.String() {
						v.D = v.String() + string(fr-r)
					}
				} else if slen == 1 && flen > 1 {
					r, _ := strconv.ParseInt(p.sv.String(), 10, 32)
					fr := rune(r)
					for _, r := range p.fv.String() {
						v.D = v.String() + string(fr-r)
					}
				} else {
					for i, r := range p.fv.String() {
						v.D = v.String() + string(r-rune(p.sv.String()[i]))
					}
				}
			default:
				fract.IPanic(p.opr, obj.ArithmeticPanic, "This operator is not defined for string types!")
			}
			return v
		}

		v.T = oop.Str
		if p.sv.T == oop.Str {
			p.fv, p.sv = p.sv, p.fv
		}
		if p.sv.T == oop.Array {
			if sl == 0 {
				v.D = p.fv.D
				return v
			}
			if len(p.fv.String()) != sl && (len(p.fv.String()) != 1 && sl != 1) {
				fract.IPanic(p.s[0], obj.ArithmeticPanic, "Array element count is not one or equals to first array!")
			}
			if strings.Contains(p.sv.String(), ".") {
				fract.IPanic(p.s[0], obj.ArithmeticPanic, "Only string and integer values can concatenate string values!")
			}
			r, _ := strconv.ParseInt(p.sv.String(), 10, 64)
			rn := rune(r)
			var sb strings.Builder
			for _, r := range p.fv.String() {
				switch p.opr.V {
				case "+":
					sb.WriteByte(byte(r + rn))
				case "-":
					sb.WriteByte(byte(r - rn))
				default:
					fract.IPanic(p.opr, obj.ArithmeticPanic, "This operator is not defined for string types!")
				}
			}
			v.D = sb.String()
		} else {
			if p.sv.T != oop.Int {
				fract.IPanic(p.s[0], obj.ArithmeticPanic, "Only string and integer values can concatenate string values!")
			}
			var s string
			rs, _ := strconv.ParseInt(p.sv.String(), 10, 64)
			rn := byte(rs)
			for _, r := range p.fv.String() {
				switch p.opr.V {
				case "+":
					s += string(byte(r) + rn)
				case "-":
					s += string(byte(r) - rn)
				default:
					fract.IPanic(p.opr, obj.ArithmeticPanic, "This operator is not defined for string types!")
				}
			}
			v.D = s
		}
		return v
	}

	if p.fv.T == oop.Array && p.sv.T == oop.Array {
		v.T = oop.Array
		if fl == 0 {
			v.D = p.sv.D
			return v
		} else if sl == 0 {
			v.D = p.fv.D
			return v
		}
		if fl != sl && fl != 1 && sl != 1 {
			fract.IPanic(p.s[0], obj.ArithmeticPanic, "Array element count is not one or equals to first array!")
		}
		if fl == 1 || sl == 1 {
			f, s := p.fv, p.sv
			if f.Len() != 1 {
				f, s = s, f
			}
			ar := str.Conv(arith(p.opr, f.D.(oop.ArrayModel)[0]))
			for i, d := range s.D.(oop.ArrayModel) {
				if d.T == oop.Array {
					s.D.(oop.ArrayModel)[i] = readyData(p, oop.Val{
						D: solveProc(process{
							f:   p.f,
							fv:  s,
							s:   p.s,
							sv:  d,
							opr: p.opr,
						}).D,
						T: oop.Array,
					})
				} else {
					s.D.(oop.ArrayModel)[i] = readyData(p, oop.Val{
						D: fmt.Sprintf(fract.FloatFormat, solve(p.opr, ar, str.Conv(arith(p.opr, d)))),
						T: oop.Int,
					})
				}
			}
			v.D = s.D
		} else {
			for i, f := range p.fv.D.(oop.ArrayModel) {
				s := p.sv.D.(oop.ArrayModel)[i]
				if f.T == oop.Array || s.T == oop.Array {
					proc := process{f: p.f, s: p.s, opr: p.opr}
					if f.T == oop.Array {
						proc.fv = oop.Val{D: f.D, T: oop.Array}
					} else {
						proc.fv = f
					}
					if s.T == oop.Array {
						proc.sv = oop.Val{D: s.D, T: oop.Array}
					} else {
						proc.sv = s
					}
					p.fv.D.(oop.ArrayModel)[i] = readyData(p, oop.Val{D: solveProc(proc).D, T: oop.Array})
				} else {
					p.fv.D.(oop.ArrayModel)[i] = readyData(p, oop.Val{
						D: fmt.Sprintf(fract.FloatFormat, solve(p.opr, str.Conv(arith(p.opr, f)), str.Conv(s.String()))),
						T: oop.Int,
					})
				}
			}
			v.D = p.fv.D
		}
	} else if p.fv.T == oop.Array || p.sv.T == oop.Array {
		v.T = oop.Array
		if p.fv.T == oop.Array && fl == 0 {
			v.D = p.sv.D
			return v
		} else if p.sv.T == oop.Array && sl == 0 {
			v.D = p.fv.D
			return v
		}
		f, s := p.fv, p.sv
		if f.T != oop.Array {
			f, s = s, f
		}
		ar := str.Conv(arith(p.opr, s))
		for i, d := range f.D.(oop.ArrayModel) {
			if d.T == oop.Array {
				f.D.(oop.ArrayModel)[i] = readyData(p, solveProc(process{
					f:   p.f,
					fv:  s,
					s:   p.s,
					sv:  d,
					opr: p.opr,
				}))
			} else {
				f.D.(oop.ArrayModel)[i] = readyData(p, oop.Val{
					D: fmt.Sprintf(fract.FloatFormat, solve(p.opr, str.Conv(arith(p.opr, d)), ar)),
					T: oop.Int,
				})
			}
		}
		v = f
	} else {
		v = readyData(p,
			oop.Val{
				D: fmt.Sprintf(fract.FloatFormat, solve(p.opr, str.Conv(arith(p.opr, p.fv)), str.Conv(arith(p.opr, p.sv)))),
				T: oop.Int,
			})
	}
	return v
}

// Select enumerable object elements.
func (p *Parser) selectEnum(mut bool, v oop.Val, tk obj.Token, s interface{}) *oop.Val {
	var r oop.Val
	switch v.T {
	case oop.Array:
		i := s.([]int)
		if len(i) == 1 {
			v := v.D.(oop.ArrayModel)[i[0]]
			if !v.Mut && !mut { //! Immutability.
				v = v.Immut()
			}
			v.Mut = v.Mut || mut
			return &v
		}
		r = oop.Val{D: oop.ArrayModel{}, T: oop.Array}
		for _, pos := range i {
			r.D = append(r.D.(oop.ArrayModel), v.D.(oop.ArrayModel)[pos])
		}
	case oop.Map:
		m := v.D.(oop.MapModel)
		switch t := s.(type) {
		case oop.ArrayModel:
			rm := oop.MapModel{}
			for _, k := range t {
				d, ok := m[k]
				if !ok {
					fract.IPanic(tk, obj.ValuePanic, "Key is not exists!")
				}
				rm[k] = d
			}
			r = oop.Val{D: rm, T: oop.Map}
		case oop.Val:
			d, ok := m[t]
			if !ok {
				fract.IPanic(tk, obj.ValuePanic, "Key is not exists!")
			}
			return &d
		}
	case oop.Str:
		r = oop.Val{D: "", T: oop.Str}
		for _, pos := range s.([]int) {
			r.D = r.String() + string(v.String()[pos])
		}
	}
	return &r
}

type valPartInfo struct {
	tks []obj.Token
	mut bool // Force to mutability.
}

// procNameVal returns value of name.
func (p *Parser) procNameVal(mut bool, tk obj.Token) *oop.Val {
	var rv *oop.Val
	vi, t := p.defByName(tk.V)
	if vi == -1 {
		if tk.V == "this" {
			fract.IPanic(tk, obj.NamePanic, `"this" keyword is cannot used this scope!`)
		}
		fract.IPanic(tk, obj.NamePanic, "Name is not defined: "+tk.V)
	}
	switch t {
	case 'f': // Function.
		rv = &oop.Val{D: p.defs.Funcs[vi], T: oop.Function}
	case 'p': // Package.
		rv = &oop.Val{D: p.packages[vi], T: oop.Package}
	case 'v': // Value.
		rv = &p.defs.Vars[vi].V
		if !rv.Mut && !mut { //! Immutability.
			*rv = rv.Immut()
		}
		rv.Mut = rv.Mut || mut
	}
	return rv
}

// Process value part.
func (p *Parser) procValPart(i valPartInfo) *oop.Val {
	var rv *oop.Val
	if i.tks[0].T == fract.Var && i.tks[0].V == "mut" {
		if len(i.tks) == 1 {
			fract.IPanic(i.tks[0], obj.SyntaxPanic, "Value is not given!")
		}
		i.mut = true
		i.tks = i.tks[1:]
		rv = p.procValPart(i)
		goto end
	}
	// Single oop.
	if tk := i.tks[0]; len(i.tks) == 1 {
		if tk.V[0] == '\'' || tk.V[0] == '"' {
			rv = &oop.Val{D: tk.V[1 : len(tk.V)-1], T: oop.Str}
			goto end
		} else if tk.V == "true" || tk.V == "false" {
			rv = &oop.Val{D: tk.V, T: oop.Bool}
			goto end
		} else if tk.T == fract.Value {
			if strings.Contains(tk.V, ".") || strings.ContainsAny(tk.V, "eE") {
				tk.T = oop.Float
			} else {
				tk.T = oop.Int
			}
			if tk.V != "NaN" {
				prs, _ := new(big.Float).SetString(tk.V)
				val, _ := prs.Float64()
				tk.V = fmt.Sprint(val)
			}
			rv = &oop.Val{D: tk.V, T: tk.T}
			goto end
		} else {
			if tk.T != fract.Name {
				fract.IPanic(tk, obj.ValuePanic, "Invalid value!")
			}
		}
	}
	switch j, tk := len(i.tks)-1, i.tks[len(i.tks)-1]; tk.T {
	case fract.Name:
		if j > 0 {
			j--
			if j == 0 || i.tks[j].T != fract.Dot {
				fract.IPanic(i.tks[j], obj.SyntaxPanic, "Invalid syntax!")
			}
			n := i.tks[j+1]
			d := i.tks[j]
			i.tks = i.tks[:j]
			v := p.procValPart(i)
			switch v.T {
			case oop.Package:
				ii := v.D.(importInfo)
				checkPublic(nil, n)
				rv = ii.src.procNameVal(i.mut, n)
				goto end
			case oop.StructIns:
				s := v.D.(oop.StructInstance)
				checkPublic(s.L, tk)
				i := s.Fields.VarIndexByName(n.V)
				if i == -1 {
					fract.IPanic(n, obj.NamePanic, "Name is not defined: "+n.V)
				}
				rv = &s.Fields.Vars[i].V
				goto end
			case oop.ClassIns:
				c := v.D.(oop.ClassInstance)
				checkPublic(c.L, tk)
				vi, t := c.Defs.DefByName(n.V)
				if vi == -1 {
					fract.IPanic(n, obj.NamePanic, "Name is not defined: "+n.V)
				}
				switch t {
				case 'f': // Function.
					rv = &oop.Val{D: c.Defs.Funcs[vi], T: oop.Function}
				case 'v': // Value.
					rv = &c.Defs.Vars[vi].V
					if !rv.Mut && !i.mut { //! Immutability.
						*rv = rv.Immut()
					}
					rv.Mut = rv.Mut || i.mut
				}
				goto end
			default:
				fract.IPanic(d, obj.ValuePanic, "Object is not support sub fields!")
			}
		}
		rv = p.procNameVal(i.mut, tk)
		goto end
	case fract.Brace:
		bc := 0
		switch tk.V {
		case ")":
			var vtks []obj.Token
			for ; j >= 0; j-- {
				t := i.tks[j]
				if t.T != fract.Brace {
					continue
				}
				switch t.V {
				case ")":
					bc++
				case "(":
					bc--
				}
				if bc > 0 {
					continue
				}
				vtks = i.tks[:j]
				break
			}
			if len(vtks) == 0 && bc == 0 {
				tk, i.tks = i.tks[0], i.tks[1:len(i.tks)-1]
				if len(i.tks) == 0 {
					fract.IPanic(tk, obj.SyntaxPanic, "Invalid syntax!")
				}
				rv = p.procVal(i.tks, i.mut)
				goto end
			}
			v := p.procValPart(valPartInfo{tks: vtks, mut: i.mut})
			switch v.T {
			case oop.Function: // Function call.
				rv = p.funcCallModel(v.D.(*oop.Func), i.tks[len(vtks):]).Call()
			case oop.StructDef:
				s := v.D.(oop.Struct)
				rv = &oop.Val{
					D: s.CallConstructor(p.funcCallModel(s.Constructor, i.tks[len(vtks):]).args),
					T: oop.StructIns,
				}
			case oop.ClassDef:
				c := v.D.(oop.Class)
				rv = &oop.Val{
					D: c.CallConstructor(p.funcCallModel(c.Constructor, i.tks[len(vtks):])),
					T: oop.ClassIns,
				}
			default:
				fract.IPanic(i.tks[len(vtks)], obj.ValuePanic, "Invalid syntax!")
			}
			goto end
		case "]":
			var vtks []obj.Token
			for ; j >= 0; j-- {
				t := i.tks[j]
				if t.T != fract.Brace {
					continue
				}
				switch t.V {
				case "]":
					bc++
				case "[":
					bc--
				}
				if bc > 0 {
					continue
				}
				vtks = i.tks[:j]
				break
			}
			if len(vtks) == 0 && bc == 0 {
				rv = p.procEnumerableVal(i.tks)
				goto end
			}
			v := p.procValPart(valPartInfo{mut: i.mut, tks: vtks})
			if v.T != oop.Array && v.T != oop.Map && v.T != oop.Str {
				fract.IPanic(vtks[0], obj.ValuePanic, "Index accessor is cannot used with not enumerable values!")
			}
			rv = p.selectEnum(i.mut, *v, tk, selections(*v, *p.procValTks(i.tks[len(vtks)+1 : len(i.tks)-1]), tk))
			goto end
		case "}":
			var vtks []obj.Token
			for ; j >= 0; j-- {
				t := i.tks[j]
				if t.T != fract.Brace {
					continue
				}
				switch t.V {
				case "}":
					bc++
				case "{":
					bc--
				}
				if bc > 0 {
					continue
				}
				vtks = i.tks[:j]
				break
			}
			l := len(vtks)
			if l == 0 && bc == 0 {
				rv = p.procEnumerableVal(i.tks)
				goto end
			} else if l > 1 && (vtks[1].T != fract.Brace || vtks[1].V != "(") {
				fract.IPanic(vtks[1], obj.SyntaxPanic, "Invalid syntax!")
			} else if l > 1 && (vtks[l-1].T != fract.Brace || vtks[l-1].V != ")") {
				fract.IPanic(vtks[l-1], obj.SyntaxPanic, "Invalid syntax!")
			}
			switch vtks[0].T {
			case fract.Func:
				f := &oop.Func{
					Name: "anonymous",
					Src:  p,
					Tks:  p.getBlock(i.tks[len(vtks):]),
				}
				if f.Tks == nil {
					f.Tks = [][]obj.Token{}
				}
				if l > 1 {
					vtks = vtks[1:]
					vtks = decomposeBrace(&vtks)
					p.setFuncParams(f, &vtks)
				}
				rv = &oop.Val{D: f, T: oop.Function}
			case fract.Struct:
				rv = p.buildStruct("anonymous", i.tks[1:])
			default:
				fract.IPanic(vtks[0], obj.SyntaxPanic, "Invalid syntax!")
			}
			vtks = nil
			goto end
		}
	}
	fract.IPanic(i.tks[0], obj.ValuePanic, "Invalid value!")
end:
	rv.Mut = i.mut
	return rv
}

// Process array oop.
func (p *Parser) procArrayVal(tks []obj.Token) *oop.Val {
	v := oop.Val{D: oop.ArrayModel{}, T: oop.Array}
	comma := 1
	bc := 0
	for j := 1; j < len(tks)-1; j++ {
		switch t := tks[j]; t.T {
		case fract.Brace:
			switch t.V {
			case "{", "[", "(":
				bc++
			default:
				bc--
			}
		case fract.Comma:
			if bc != 0 {
				break
			}
			if comma-j == 0 {
				fract.IPanic(t, obj.SyntaxPanic, "Value is not given!")
			}
			v.D = append(v.D.(oop.ArrayModel), *p.procValTks(tks[comma:j]))
			comma = j + 1
		}
	}
	if l := len(tks); comma < l-1 {
		v.D = append(v.D.(oop.ArrayModel), *p.procValTks(tks[comma : l-1]))
	}
	return &v
}

// Process map oop.
func (p *Parser) procMapVal(tks []obj.Token) *oop.Val {
	comma := 1
	bc := 0
	m := oop.MapModel{}
	for j := 1; j < len(tks)-1; j++ {
		switch t := tks[j]; t.T {
		case fract.Brace:
			switch t.V {
			case "{", "[", "(":
				bc++
			default:
				bc--
			}
		case fract.Comma:
			if bc != 0 {
				break
			}
			if comma-j == 0 {
				fract.IPanic(t, obj.SyntaxPanic, "Value is not given!")
			}
			lst := tks[comma:j]
			var (
				i  int
				l  int = len(lst)
				tk obj.Token
			)
			for i, tk = range lst {
				switch tk.T {
				case fract.Brace:
					switch tk.V {
					case "{", "[", "(":
						bc++
					default:
						bc--
					}
				case fract.Colon:
					if bc != 0 {
						break
					}
					if i+1 >= l {
						fract.IPanic(tk, obj.SyntaxPanic, "Value is not given!")
					}
					key := *p.procValTks(lst[:i])
					_, ok := m[key]
					if ok {
						fract.IPanic(tk, obj.ValuePanic, "Key is already defined!")
					}
					m[key] = *p.procValTks(lst[i+1:])
					comma = j + 1
					lst = nil
				}
			}
			if lst != nil {
				fract.IPanic(lst[l-1], obj.SyntaxPanic, "Value identifier is not found!")
			}
		}
	}
	if comma < len(tks)-1 {
		lst := tks[comma : len(tks)-1]
		i := -1
		l := len(lst)
		for j, tk := range lst {
			switch tk.T {
			case fract.Brace:
				switch tk.V {
				case "{", "[", "(":
					bc++
				default:
					bc--
				}
			case fract.Colon:
				if bc != 0 {
					break
				}
				i = j
			}
			if i != -1 {
				break
			}
		}
		if i+1 >= l {
			fract.IPanic(lst[i], obj.SyntaxPanic, "Value is not given!")
		}
		key := *p.procValTks(lst[:i])
		_, ok := m[key]
		if ok {
			fract.IPanic(lst[i], obj.ValuePanic, "Key is already defined!")
		}
		m[key] = *p.procValTks(lst[i+1:])
		lst = nil
	}
	return &oop.Val{D: m, T: oop.Map}
}

// Process list comprehension.
func (p *Parser) procListComprehension(tks []obj.Token) *oop.Val {
	var (
		stks []obj.Token // Select tokens.
		ltks []obj.Token // Loop tokens.
		ftks []obj.Token // Filter tokens.
		bc   int
	)
	for i, t := range tks {
		if t.T == fract.Brace {
			switch t.V {
			case "{", "[", "(":
				bc++
			default:
				bc--
			}
		}
		if bc > 1 {
			continue
		}
		if t.T == fract.Loop {
			stks = tks[1:i]
		} else if t.T == fract.Comma {
			ltks = tks[len(stks)+1 : i]
			ftks = tks[i+1 : len(tks)-1]
			if len(ftks) == 0 {
				ftks = nil
			}
			break
		}
	}
	if ltks == nil {
		ltks = tks[len(stks)+1 : len(tks)-1]
	}
	if len(ltks) < 2 {
		fract.IPanic(ltks[0], obj.SyntaxPanic, "Variable name is not given!")
	}
	nametk := ltks[1]
	// Name is not name?
	if nametk.T != fract.Name {
		fract.IPanic(nametk, obj.SyntaxPanic, "This is not a valid name!")
	}
	if ln := p.definedName(nametk.V); ln != -1 {
		fract.IPanic(nametk, obj.NamePanic, "\""+nametk.V+"\" is already defined at line: "+fmt.Sprint(ln))
	}
	if l := len(ltks); l < 3 {
		tk := tks[0]
		fract.IPanicC(tk.F, tk.Ln, ltks[1].Col+len(ltks[1].V), obj.SyntaxPanic, "Value is not given!")
	} else if l < 4 {
		fract.IPanic(ltks[2], obj.SyntaxPanic, "Value is not given!")
	}
	ltks = ltks[3:]
	varr := *p.procValTks(ltks)
	// Type is not array?
	if !varr.IsEnum() {
		fract.IPanic(ltks[0], obj.ValuePanic, "Foreach loop must defined enumerable value!")
	}
	element := &oop.Var{Name: nametk.V}
	if element.Name == "_" {
		element.Name = ""
	}
	p.defs.Vars = append(p.defs.Vars, element)
	// Interpret block.
	v := oop.Val{D: oop.ArrayModel{}, T: oop.Array}
	l := loop{enum: varr}
	l.run(func() {
		element.V = l.b
		if ftks == nil || p.procCondition(ftks) == "true" {
			v.D = append(v.D.(oop.ArrayModel), *p.procValTks(stks))
		}
	})
	// Remove variables.
	element = nil
	p.defs.Vars = p.defs.Vars[:len(p.defs.Vars)-1]
	return &v
}

// Process enumerable oop.
func (p *Parser) procEnumerableVal(tks []obj.Token) *oop.Val {
	var (
		lc bool
		bc int
	)
	for _, t := range tks {
		if t.T == fract.Brace {
			switch t.V {
			case "{", "[", "(":
				bc++
			default:
				bc--
			}
		}
		if bc > 1 {
			continue
		}
		if t.T == fract.Comma {
			break
		} else if !lc && t.T == fract.Loop {
			if tks[0].V != "[" {
				fract.IPanic(tks[0], obj.SyntaxPanic, "Invalid syntax!")
			}
			lc = true
			break
		}
	}
	if lc {
		return p.procListComprehension(tks)
	} else if tks[0].V == "{" {
		return p.procMapVal(tks)
	}
	return p.procArrayVal(tks)
}

// Process oop.
func (p *Parser) procVal(tks []obj.Token, mut bool) *oop.Val {
	// Is conditional expression?
	if j, _ := findConditionOpr(tks); j != -1 {
		return &oop.Val{D: p.procCondition(tks), T: oop.Bool}
	}
	procs := arithmeticProcesses(tks)
	i := valPartInfo{mut: mut}
	if len(procs) == 1 {
		i.tks = procs[0]
		return p.procValPart(i)
	}
	var v oop.Val
	var opr process
	j := nextopr(procs)
	for j != -1 {
		if j == 0 {
			if len(procs) == 1 {
				break
			}
			opr.fv = v
			opr.opr = procs[j][0]
			opr.s = procs[j+1]
			i.tks = opr.s
			opr.sv = *p.procValPart(i)
			if opr.sv.T == fract.None {
				fract.IPanic(opr.f[0], obj.ValuePanic, "Value is not given!")
			}
			v = solveProc(opr)
			procs = procs[2:]
			j = nextopr(procs)
			continue
		}
		opr.f = procs[j-1]
		i.tks = opr.f
		opr.fv = *p.procValPart(i)
		if opr.fv.T == fract.None {
			fract.IPanic(opr.f[0], obj.ValuePanic, "Value is not given!")
		}
		opr.opr = procs[j][0]
		opr.s = procs[j+1]
		i.tks = opr.s
		opr.sv = *p.procValPart(i)
		if opr.sv.T == fract.None {
			fract.IPanic(opr.s[0], obj.ValuePanic, "Value is not given!")
		}
		rv := solveProc(opr)
		if v.D != nil {
			opr.opr.V = "+"
			opr.s = procs[j+1]
			opr.fv = v
			opr.sv = rv
			v = solveProc(opr)
		} else {
			v = rv
		}
		// Remove computed processes.
		procs = append(procs[:j-1], procs[j+2:]...)
		// Find next operator.
		j = nextopr(procs)
	}
	procs = nil
	opr.f = nil
	opr.s = nil
	return &v
}

// Process value from tokens.
func (p *Parser) procValTks(tks []obj.Token) *oop.Val { return p.procVal(tks, false) }
