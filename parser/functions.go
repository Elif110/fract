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
	fn    *oop.Fn
	errTk obj.Token
	args  []oop.VarDef
}

func (c *funcCall) Func() *oop.Fn { return c.fn }

func (c *funcCall) Call() *oop.Val {
	var returnVal oop.Val
	// Is built-in function?
	if c.fn.Tokens == nil {
		returnVal = c.fn.Src.(func(obj.Token, []oop.VarDef) oop.Val)(c.errTk, c.args)
		c.args = nil
		c.fn = nil
		return &returnVal
	}
	// Process block.
	deferLen := len(defers)
	src := c.fn.Src.(*Parser)
	p := Parser{
		defs:         oop.DefMap{Vars: append(c.args, c.fn.Args...), Funcs: src.defs.Funcs},
		packages:     src.packages,
		funcTempVars: src.funcTempVars,
		loopCount:    0,
		funcCount:    1,
		Tokens:       c.fn.Tokens[:len(c.fn.Tokens):len(c.fn.Tokens)],
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
	block := obj.Block{
		Try: func() {
			for p.index = 0; p.index < len(p.Tokens); p.index++ {
				if p.processExpression(p.Tokens[p.index]) == fract.FUNCReturn {
					src.returnVal = p.returnVal
					if src.returnVal == nil {
						break
					}
					returnVal = *src.returnVal
					src.returnVal = nil
					break
				}
			}
		},
	}
	block.Do()
	if block.Panic.Msg != "" {
		defers = defers[:deferLen]
		panic(block.Panic.Msg)
	}
	for i := len(defers) - 1; i >= deferLen; i-- {
		defers[i].Call()
	}
	defers = defers[:deferLen]
	c.args = nil
	c.fn = nil
	return &returnVal
}

// isParamSet Argument type is param set?
func isParamSet(tokens []obj.Token) bool {
	return len(tokens) >= 2 && tokens[0].Type == fract.Name && tokens[1].Val == "="
}

// paramsArgValues decompose and returns params values.
func (p *Parser) paramsArgValues(tokens []obj.Token, index, lastComma *int, mut bool) oop.Val {
	values := oop.NewListModel()
	resultVal := oop.Val{Type: oop.List}
	braceCount := 0
	for ; *index < len(tokens); *index++ {
		switch tk := tokens[*index]; tk.Type {
		case fract.Brace:
			switch tk.Val {
			case "{", "[", "(":
				braceCount++
			default:
				braceCount--
			}
		case fract.Comma:
			if braceCount != 0 {
				break
			}
			valTokens := tokens[*lastComma:*index]
			if isParamSet(valTokens) {
				*index -= 4
				resultVal.Data = values
				return resultVal
			}
			var params bool
			if l := len(valTokens); valTokens[l-1].Type == fract.Params {
				tk = valTokens[l-1]
				params = true
				valTokens = valTokens[:l-1]
			}
			val := *p.processValue(valTokens, mut)
			if params {
				if val.Type != oop.List {
					fract.IPanic(tk, obj.ValuePanic, "Notation is can used for only lists!")
				}
				values.PushBack(val.Data.(*oop.ListModel).Elems...)
			} else {
				values.PushBack(val)
			}
			valTokens = nil
			*lastComma = *index + 1
		}
	}
	if *lastComma < len(tokens) {
		valTokens := tokens[*lastComma:]
		if isParamSet(valTokens) {
			*index -= 4
			return resultVal
		}
		var params bool
		var tk obj.Token
		if l := len(valTokens); valTokens[l-1].Type == fract.Params {
			tk = valTokens[l-1]
			params = true
			valTokens = valTokens[:l-1]
		}
		val := *p.processValue(valTokens, mut)
		if params {
			if val.Type != oop.List {
				fract.IPanic(tk, obj.ValuePanic, "Notation is can used for only lists!")
			}
			values.PushBack(val.Data.(*oop.ListModel).Elems...)
		} else {
			values.PushBack(val)
		}
		valTokens = nil
	}
	resultVal.Data = values
	return resultVal
}

type funcArgInfo struct {
	fn        *oop.Fn
	names     *[]string
	tokens    []obj.Token
	tk        obj.Token
	index     *int
	count     *int
	lastComma *int
}

// Process function argument.
func (p *Parser) processFuncArg(inf funcArgInfo) *oop.Var {
	var paramSet bool
	length := *inf.index - *inf.lastComma
	if length < 1 {
		fract.IPanic(inf.tk, obj.SyntaxPanic, "Value is not given!")
	} else if *inf.count >= len(inf.fn.Params) {
		fract.IPanic(inf.tk, obj.SyntaxPanic, "Argument overflow!")
	}
	param := inf.fn.Params[*inf.count]
	resultVar := &oop.Var{Name: param.Name}
	valTokens := inf.tokens[*inf.lastComma:*inf.index]
	inf.tk = valTokens[0]
	// Check param set.
	if length >= 2 && isParamSet(valTokens) {
		length -= 2
		if length < 1 {
			fract.IPanic(inf.tk, obj.SyntaxPanic, "Value is not given!")
		}
		for _, param := range inf.fn.Params {
			if param.Name == inf.tk.Val {
				for _, name := range *inf.names {
					if name == inf.tk.Val {
						fract.IPanic(inf.tk, obj.SyntaxPanic, "Keyword argument repeated!")
					}
				}
				*inf.count++
				paramSet = true
				*inf.names = append(*inf.names, inf.tk.Val)
				resultVar.Name = inf.tk.Val
				//Parameter is params typed?
				if param.Params {
					*inf.lastComma += 2
					resultVar.Val = p.paramsArgValues(inf.tokens, inf.index, inf.lastComma, param.Type == "mut")
				} else {
					resultVar.Val = *p.processValue(valTokens[2:], param.Type == "mut")
				}
				resultVar.Val.Const = param.Type == "const"
				return resultVar
			}
		}
		fract.IPanic(inf.tk, obj.NamePanic, "Parameter is not defined in this name: "+inf.tk.Val)
	}
	if paramSet {
		fract.IPanic(inf.tk, obj.SyntaxPanic, "After the parameter has been given a special value, all parameters must be shown privately!")
	}
	*inf.count++
	*inf.names = append(*inf.names, resultVar.Name)
	// Parameter is params typed?
	if param.Params {
		resultVar.Val = p.paramsArgValues(inf.tokens, inf.index, inf.lastComma, param.Type == "mut")
	} else {
		resultVar.Val = *p.processValue(valTokens, param.Type == "mut")
	}
	resultVar.Val.Const = param.Type == "const"
	valTokens = nil
	return resultVar
}

// Process function call model and initialize model instance.
func (p *Parser) funcCallModel(fn *oop.Fn, tokens []obj.Token) *funcCall {
	var (
		names    []string
		args     []oop.VarDef
		argCount = 0
		tk       = tokens[0]
	)
	// Decompose arguments.
	tokens = decomposeBrace(&tokens)
	var (
		inf = funcArgInfo{
			fn:        fn,
			names:     &names,
			tk:        tk,
			tokens:    tokens,
			count:     &argCount,
			index:     new(int),
			lastComma: new(int),
		}
		braceCount = 0
	)
	for *inf.index = 0; *inf.index < len(tokens); *inf.index++ {
		switch inf.tk = tokens[*inf.index]; inf.tk.Type {
		case fract.Brace:
			switch inf.tk.Val {
			case "{", "[", "(":
				braceCount++
			default:
				braceCount--
			}
		case fract.Comma:
			if braceCount != 0 {
				break
			}
			args = append(args, p.processFuncArg(inf))
			*inf.lastComma = *inf.index + 1
		}
	}
	if *inf.lastComma < len(tokens) {
		inf.tk = tokens[*inf.lastComma]
		tkslen := len(tokens)
		inf.index = &tkslen
		args = append(args, p.processFuncArg(inf))
	}
	tokens = nil
	inf.count = nil
	inf.index = nil
	inf.lastComma = nil
	inf.names = nil
	// All parameters is not defined?
	if argCount < len(fn.Params)-fn.DefaultParamCount {
		var sb strings.Builder
		sb.WriteString("All required positional arguments is not given:")
		for _, p := range fn.Params {
			if p.DefaultVal.Data != nil {
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
	for ; argCount < len(fn.Params); argCount++ {
		param := fn.Params[argCount]
		if param.DefaultVal.Data != nil {
			args = append(args, &oop.Var{Name: param.Name, Val: param.DefaultVal})
		}
	}
	return &funcCall{fn: fn, errTk: tk, args: args}
}

// Set arguments to parameters of function.
func (p *Parser) setParams(fn *oop.Fn, tokens *[]obj.Token) {
	paramName, params, defaultDef, varType := true, false, false, ""
	braceCount := 0
	var param oop.Param
	for i := 0; i < len(*tokens); i++ {
		tk := (*tokens)[i]
		if tk.Type == fract.Brace {
			switch tk.Val {
			case "{", "[", "(":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		}
		if paramName {
			switch tk.Type {
			case fract.Params:
				if params {
					fract.IPanic(tk, obj.SyntaxPanic, "Invalid syntax!")
				}
				params = true
				continue
			case fract.Name:
				if !isValidName(tk.Val) {
					fract.IPanic(tk, obj.NamePanic, "Invalid name!")
				}
			case fract.Var:
				if varType != "" || params {
					fract.IPanic(tk, obj.SyntaxPanic, "Invalid syntax!")
				}
				varType = tk.Val
				continue
			default:
				fract.IPanic(tk, obj.SyntaxPanic, "Parameter name is not found!")
			}
			param = oop.Param{Name: tk.Val, Params: params, Type: varType}
			fn.Params = append(fn.Params, param)
			paramName = false
			varType = ""
			continue
		} else {
			paramName = true
			// Default value definition?
			if tk.Val == "=" {
				braceCount := 0
				i++
				start := i
				for ; i < len(*tokens); i++ {
					tk = (*tokens)[i]
					if tk.Type == fract.Brace {
						switch tk.Val {
						case "{", "[", "(":
							braceCount++
						default:
							braceCount--
						}
					} else if tk.Type == fract.Comma {
						break
					}
				}
				if i-start < 1 {
					fract.IPanic((*tokens)[start-1], obj.SyntaxPanic, "Value is not given!")
				}
				param.DefaultVal = *p.processValTokens((*tokens)[start:i])
				if param.Params && param.DefaultVal.Type != oop.List {
					fract.IPanic(tk, obj.ValuePanic, "Params parameter is can only take list values!")
				}
				fn.Params[len(fn.Params)-1] = param
				fn.DefaultParamCount++
				defaultDef = true
				continue
			}
			if param.DefaultVal.Data == nil && defaultDef {
				fract.IPanic(tk, obj.SyntaxPanic, "All parameters after a given parameter with a default value must take a default value!")
			} else if tk.Type != fract.Comma {
				fract.IPanic(tk, obj.SyntaxPanic, "Comma is not found!")
			}
		}
	}
	if param.DefaultVal.Data == nil && defaultDef {
		fract.IPanic((*tokens)[len(*tokens)-1], obj.SyntaxPanic, "All parameters after a given parameter with a default value must take a default value!")
	}
}

// Process function declaration to defmap.
func (p *Parser) ffuncdec(defs *oop.DefMap, tokens []obj.Token) {
	tokensLen := len(tokens)
	nameTk := tokens[1]
	// Name is not name?
	if nameTk.Type != fract.Name || !isValidName(nameTk.Val) {
		fract.IPanic(nameTk, obj.SyntaxPanic, "Invalid name!")
	}
	// Name is already defined?
	var ln int
	if &p.defs == defs { // Variable added to defmap of parser.
		ln = p.defLineByName(nameTk.Val)
	} else { // Variable added to another defmap.
		ln = defs.DefIndexByName(nameTk.Val)
	}
	if ln != -1 {
		fract.IPanic(nameTk, obj.NamePanic, "\""+nameTk.Val+"\" is already defined at line: "+fmt.Sprint(ln))
	}

	if tokensLen < 3 {
		fract.IPanicC(nameTk.File, nameTk.Line, nameTk.Column+len(nameTk.Val), obj.SyntaxPanic, "Invalid syntax!")
	}
	fn := &oop.Fn{
		Name: nameTk.Val,
		Line: p.index,
		Src:  p,
	}
	// Decompose function arguments.
	if tokens[2].Val == "(" {
		tokens = tokens[2:]
		r := decomposeBrace(&tokens)
		p.setParams(fn, &r)
		r = nil
	} else {
		tokens = tokens[2:]
	}
	fn.Tokens = p.getBlock(tokens)
	if fn.Tokens == nil {
		fn.Tokens = [][]obj.Token{}
	}
	fn.Line = nameTk.Line
	defs.Funcs = append(defs.Funcs, fn)
}

// Process function declaration to defmap of parser.
func (p *Parser) funcdec(tks []obj.Token) { p.ffuncdec(&p.defs, tks) }
