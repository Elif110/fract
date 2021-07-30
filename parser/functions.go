package parser

import (
	"fmt"
	"strings"

	"github.com/fract-lang/fract/oop"
	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
)

// Instance for function calls.
type funcCall struct {
	f     *oop.Func
	errTk obj.Token
	args  []*oop.Var
}

func (c *funcCall) Func() *oop.Func { return c.f }

func (c *funcCall) Call() *oop.Val {
	var retv oop.Val
	// Is built-in function?
	if c.f.Tks == nil {
		retv = c.f.Src.(func(tk obj.Token, args []*oop.Var) oop.Val)(c.errTk, c.args)
		return &retv
	}
	// Process block.
	dlen := len(defers)
	src := c.f.Src.(*Parser)
	p := Parser{
		defs:         oop.DefMap{Vars: append(c.args, c.f.Args...), Funcs: src.defs.Funcs},
		packages:     src.packages,
		funcTempVars: src.funcTempVars,
		loopCount:    0,
		funcCount:    1,
		Tks:          c.f.Tks[:len(c.f.Tks):len(c.f.Tks)],
	}
	if p.funcTempVars == -1 {
		p.funcTempVars = 0
	}
	if p.funcTempVars == 0 {
		p.defs.Vars = append(p.defs.Vars, src.defs.Vars...)
	} else {
		p.defs.Vars = append(p.defs.Vars, src.defs.Vars[:len(src.defs.Vars)-p.funcTempVars]...)
	}
	p.funcTempVars = len(c.args)
	// Interpret block.
	b := obj.Block{
		Try: func() {
			for p.i = 0; p.i < len(p.Tks); p.i++ {
				if p.process(p.Tks[p.i]) == fract.FUNCReturn {
					src.retVal = p.retVal
					if src.retVal == nil {
						break
					}
					retv = *src.retVal
					src.retVal = nil
					break
				}
			}
		},
	}
	b.Do()
	if b.P.M != "" {
		defers = defers[:dlen]
		panic(b.P.M)
	}
	for i := len(defers) - 1; i >= dlen; i-- {
		defers[i].Call()
	}
	defers = defers[:dlen]
	return &retv
}

// isParamSet Argument type is param set?
func isParamSet(tks []obj.Token) bool {
	return len(tks) >= 2 && tks[0].T == fract.Name && tks[1].V == "="
}

// paramsArgVals decompose and returns params values.
func (p *Parser) paramsArgVals(tks []obj.Token, i, lstComma *int) oop.Val {
	var data oop.ArrayModel
	retv := oop.Val{T: oop.Array}
	bc := 0
	for ; *i < len(tks); *i++ {
		switch tk := tks[*i]; tk.T {
		case fract.Brace:
			switch tk.V {
			case "{", "[", "(":
				bc++
			default:
				bc--
			}
		case fract.Comma:
			if bc != 0 {
				break
			}
			vtks := tks[*lstComma:*i]
			if isParamSet(vtks) {
				*i -= 4
				retv.D = data
				return retv
			}
			var params bool
			if l := len(vtks); vtks[l-1].T == fract.Params {
				tk = vtks[l-1]
				params = true
				vtks = vtks[:l-1]
			}
			v := *p.procValTks(vtks)
			if params {
				if v.T != oop.Array {
					fract.IPanic(tk, obj.ValuePanic, "Notation is can used for only arrays!")
				}
				data = append(data, v.D.(oop.ArrayModel)...)
			} else {
				data = append(data, v)
			}
			vtks = nil
			*lstComma = *i + 1
		}
	}
	if *lstComma < len(tks) {
		vtks := tks[*lstComma:]
		if isParamSet(vtks) {
			*i -= 4
			return retv
		}
		var params bool
		var tk obj.Token
		if l := len(vtks); vtks[l-1].T == fract.Params {
			tk = vtks[l-1]
			params = true
			vtks = vtks[:l-1]
		}
		v := *p.procValTks(vtks)
		if params {
			if v.T != oop.Array {
				fract.IPanic(tk, obj.ValuePanic, "Notation is can used for only arrays!")
			}
			data = append(data, v.D.(oop.ArrayModel)...)
		} else {
			data = append(data, v)
		}
		vtks = nil
	}
	retv.D = data
	return retv
}

type funcArgInfo struct {
	f        *oop.Func
	names    *[]string
	tks      []obj.Token
	tk       obj.Token
	index    *int
	count    *int
	lstComma *int
}

// Process function argument.
func (p *Parser) procFuncArg(i funcArgInfo) *oop.Var {
	var paramSet bool
	l := *i.index - *i.lstComma
	if l < 1 {
		fract.IPanic(i.tk, obj.SyntaxPanic, "Value is not given!")
	} else if *i.count >= len(i.f.Params) {
		fract.IPanic(i.tk, obj.SyntaxPanic, "Argument overflow!")
	}
	param := i.f.Params[*i.count]
	v := &oop.Var{Name: param.Name}
	vtks := i.tks[*i.lstComma:*i.index]
	i.tk = vtks[0]
	// Check param set.
	if l >= 2 && isParamSet(vtks) {
		l -= 2
		if l < 1 {
			fract.IPanic(i.tk, obj.SyntaxPanic, "Value is not given!")
		}
		for _, pr := range i.f.Params {
			if pr.Name == i.tk.V {
				for _, name := range *i.names {
					if name == i.tk.V {
						fract.IPanic(i.tk, obj.SyntaxPanic, "Keyword argument repeated!")
					}
				}
				*i.count++
				paramSet = true
				*i.names = append(*i.names, i.tk.V)
				retv := &oop.Var{Name: i.tk.V}
				//Parameter is params typed?
				if pr.Params {
					*i.lstComma += 2
					retv.V = p.paramsArgVals(i.tks, i.index, i.lstComma)
				} else {
					retv.V = *p.procValTks(vtks[2:])
				}
				return retv
			}
		}
		fract.IPanic(i.tk, obj.NamePanic, "Parameter is not defined in this name: "+i.tk.V)
	}
	if paramSet {
		fract.IPanic(i.tk, obj.SyntaxPanic, "After the parameter has been given a special value, all parameters must be shown privately!")
	}
	*i.count++
	*i.names = append(*i.names, v.Name)
	// Parameter is params typed?
	if param.Params {
		v.V = p.paramsArgVals(i.tks, i.index, i.lstComma)
	} else {
		v.V = *p.procValTks(vtks)
	}
	vtks = nil
	return v
}

// Process function call model and initialize model instance.
func (p *Parser) funcCallModel(f *oop.Func, tks []obj.Token) *funcCall {
	var (
		names []string
		args  []*oop.Var
		count = 0
		tk    = tks[0]
	)
	// Decompose arguments.
	tks = decomposeBrace(&tks)
	var (
		inf = funcArgInfo{
			f:        f,
			names:    &names,
			tk:       tk,
			tks:      tks,
			count:    &count,
			index:    new(int),
			lstComma: new(int),
		}
		bc = 0
	)
	for *inf.index = 0; *inf.index < len(tks); *inf.index++ {
		switch inf.tk = tks[*inf.index]; inf.tk.T {
		case fract.Brace:
			switch inf.tk.V {
			case "{", "[", "(":
				bc++
			default:
				bc--
			}
		case fract.Comma:
			if bc != 0 {
				break
			}
			args = append(args, p.procFuncArg(inf))
			*inf.lstComma = *inf.index + 1
		}
	}
	if *inf.lstComma < len(tks) {
		inf.tk = tks[*inf.lstComma]
		tkslen := len(tks)
		inf.index = &tkslen
		args = append(args, p.procFuncArg(inf))
	}
	tks = nil
	inf.count = nil
	inf.index = nil
	inf.lstComma = nil
	inf.names = nil
	// All parameters is not defined?
	if count < len(f.Params)-f.DefParamCount {
		var sb strings.Builder
		sb.WriteString("All required positional arguments is not given:")
		for _, p := range f.Params {
			if p.Defval.D != nil {
				break
			}
			msg := " '" + p.Name + "',"
			for _, name := range names {
				if p.Name == name {
					msg = ""
					break
				}
			}
			sb.WriteString(msg)
		}
		fract.IPanic(tk, obj.PlainPanic, sb.String()[:sb.Len()-1])
	}
	// Check default values.
	for ; count < len(f.Params); count++ {
		p := f.Params[count]
		if p.Defval.D != nil {
			args = append(args, &oop.Var{Name: p.Name, V: p.Defval})
		}
	}
	return &funcCall{
		f:     f,
		errTk: tk,
		args:  args,
	}
}

// Decompose function parameters.
func (p *Parser) setFuncParams(f *oop.Func, tks *[]obj.Token) {
	pname, params, defaultDef := true, false, false
	bc := 0
	var lstp oop.Param
	for i := 0; i < len(*tks); i++ {
		pr := (*tks)[i]
		if pr.T == fract.Brace {
			switch pr.V {
			case "{", "[", "(":
				bc++
			default:
				bc--
			}
		}
		if bc > 0 {
			continue
		}
		if pname {
			switch pr.T {
			case fract.Params:
				if params {
					fract.IPanic(pr, obj.SyntaxPanic, "Invalid syntax!")
				}
				params = true
				continue
			case fract.Name:
				if !validName(pr.V) {
					fract.IPanic(pr, obj.NamePanic, "Invalid name!")
				}
			default:
				fract.IPanic(pr, obj.SyntaxPanic, "Parameter name is not found!")
			}
			lstp = oop.Param{Name: pr.V, Params: params}
			f.Params = append(f.Params, lstp)
			pname = false
			continue
		} else {
			pname = true
			// Default value definition?
			if pr.V == "=" {
				bc := 0
				i++
				start := i
				for ; i < len(*tks); i++ {
					pr = (*tks)[i]
					if pr.T == fract.Brace {
						switch pr.V {
						case "{", "[", "(":
							bc++
						default:
							bc--
						}
					} else if pr.T == fract.Comma {
						break
					}
				}
				if i-start < 1 {
					fract.IPanic((*tks)[start-1], obj.SyntaxPanic, "Value is not given!")
				}
				lstp.Defval = *p.procValTks((*tks)[start:i])
				if lstp.Params && lstp.Defval.T != oop.Array {
					fract.IPanic(pr, obj.ValuePanic, "Params parameter is can only take array values!")
				}
				f.Params[len(f.Params)-1] = lstp
				f.DefParamCount++
				defaultDef = true
				continue
			}
			if lstp.Defval.D == nil && defaultDef {
				fract.IPanic(pr, obj.SyntaxPanic, "All parameters after a given parameter with a default value must take a default value!")
			} else if pr.T != fract.Comma {
				fract.IPanic(pr, obj.SyntaxPanic, "Comma is not found!")
			}
		}
	}
	if lstp.Defval.D == nil && defaultDef {
		fract.IPanic((*tks)[len(*tks)-1], obj.SyntaxPanic, "All parameters after a given parameter with a default value must take a default value!")
	}
}

// Process function declaration to defmap.
func (p *Parser) ffuncdec(dm *oop.DefMap, tks []obj.Token) {
	tkslen := len(tks)
	name := tks[1]
	// Name is not name?
	if name.T != fract.Name || !validName(name.V) {
		fract.IPanic(name, obj.SyntaxPanic, "Invalid name!")
	}
	// Name is already defined?
	var ln int
	if &p.defs == dm { // Variable added to defmap of parser.
		ln = p.definedName(name.V)
	} else { // Variable added to another defmap.
		ln = dm.DefinedName(name.V)
	}
	if ln != -1 {
		fract.IPanic(name, obj.NamePanic, "\""+name.V+"\" is already defined at line: "+fmt.Sprint(ln))
	}

	if tkslen < 3 {
		fract.IPanicC(name.F, name.Ln, name.Col+len(name.V), obj.SyntaxPanic, "Invalid syntax!")
	}
	f := &oop.Func{
		Name: name.V,
		Ln:   p.i,
		Src:  p,
	}
	// Decompose function parameters.
	if tks[2].V == "(" {
		tks = tks[2:]
		r := decomposeBrace(&tks)
		p.setFuncParams(f, &r)
		r = nil
	} else {
		tks = tks[2:]
	}
	f.Tks = p.getBlock(tks)
	if f.Tks == nil {
		f.Tks = [][]obj.Token{}
	}
	f.Ln = name.Ln
	dm.Funcs = append(dm.Funcs, f)
}

// Process function declaration to defmap of parser.
func (p *Parser) funcdec(tks []obj.Token) { p.ffuncdec(&p.defs, tks) }
