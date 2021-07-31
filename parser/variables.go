package parser

import (
	"fmt"

	"github.com/fract-lang/fract/oop"
	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
)

// Metadata of variable declaration.
type varInfo struct {
	shortDeclaration bool
	constant         bool
	mut              bool
}

// Append variable to source.
func (p *Parser) varadd(defs *oop.DefMap, inf varInfo, tokens []obj.Token) {
	nameTk := tokens[0]
	if !isValidName(nameTk.Val) {
		fract.IPanic(nameTk, obj.SyntaxPanic, "Invalid name!")
	}
	// Name is already defined?
	var ln int
	if &p.defs == defs { // Variable added to defmap of parser.
		ln = p.defIndexByName(nameTk.Val)
	} else { // Variable added to another defmap.
		ln = defs.DefIndexByName(nameTk.Val)
	}
	if ln != -1 {
		fract.IPanic(nameTk, obj.NamePanic, "\""+nameTk.Val+"\" is already defined at line: "+fmt.Sprint(ln))
	}

	tokensLen := len(tokens)
	// Setter is not defined?
	if tokensLen < 2 {
		fract.IPanicC(nameTk.File, nameTk.Line, nameTk.Column+len(nameTk.Val), obj.SyntaxPanic, "Setter is not found!")
	}
	setter := tokens[1]
	// Setter is not a setter operator?
	if setter.Type != fract.Operator || (setter.Val != "=" && !inf.shortDeclaration || setter.Val != ":=" && inf.shortDeclaration) {
		fract.IPanic(setter, obj.SyntaxPanic, "Invalid setter operator: "+setter.Val)
	}
	// Value is not defined?
	if tokensLen < 3 {
		fract.IPanicC(setter.File, setter.Line, setter.Column+len(setter.Val), obj.SyntaxPanic, "Value is not given!")
	}
	val := *p.processValTokens(tokens[2:])
	if val.Data == nil {
		fract.IPanic(tokens[2], obj.ValuePanic, "Invalid value!")
	}
	if p.funcTempVars != -1 {
		p.funcTempVars++
	}
	val.Mut = inf.mut
	val.Const = inf.constant
	defs.Vars = append(defs.Vars, oop.Var{
		Name: nameTk.Val,
		Val:  val,
		Line: nameTk.Line,
	})
}

// Process variable declaration to defmap.
func (p *Parser) fvardec(defs *oop.DefMap, tokens []obj.Token) {
	// Name is not defined?
	if len(tokens) < 2 {
		first := tokens[0]
		fract.IPanicC(first.File, first.Line, first.Column+len(first.Val), obj.SyntaxPanic, "Name is not given!")
	}
	inf := varInfo{
		constant: tokens[0].Val == "const",
		mut:      tokens[0].Val == "mut",
	}
	pre := tokens[1]
	if pre.Type == fract.Name {
		p.varadd(defs, inf, tokens[1:])
	} else if pre.Type == fract.Brace && pre.Val == "(" {
		tokens = tokens[2 : len(tokens)-1]
		last := 0
		line := tokens[0].Line
		braceCount := 0
		for j, tk := range tokens {
			if tk.Type == fract.Brace {
				switch tk.Val {
				case "{", "[", "(":
					braceCount++
				default:
					braceCount--
					line = tk.Line
				}
			}
			if braceCount > 0 {
				continue
			}
			if line < tk.Line {
				p.varadd(defs, inf, tokens[last:j])
				last = j
				line = tk.Line
			}
		}
		if len(tokens) != last {
			p.varadd(defs, inf, tokens[last:])
		}
	} else {
		fract.IPanic(pre, obj.SyntaxPanic, "Invalid syntax!")
	}
}

// Process variable declaration to parser.
func (p *Parser) vardec(tokens []obj.Token) { p.fvardec(&p.defs, tokens) }

// Process short variable declaration.
func (p *Parser) varsdec(tokens []obj.Token) {
	// Name is not defined?
	if len(tokens) < 2 {
		first := tokens[0]
		fract.IPanicC(first.File, first.Line, first.Column+len(first.Val), obj.SyntaxPanic, "Name is not given!")
	}
	if tokens[0].Type != fract.Name {
		fract.IPanic(tokens[0], obj.SyntaxPanic, "Invalid syntax!")
	}
	var inf varInfo
	inf.shortDeclaration = true
	p.varadd(&p.defs, inf, tokens)
}

// Process variable set statement.
func (p *Parser) varset(tokens []obj.Token) {
	var (
		enumVal    *oop.Val
		selections interface{}
		valTokens  []obj.Token
		setter     obj.Token
	)
	braceCount := 0
	lastOpenBrace := -1
	for i, tk := range tokens {
		if tk.Type == fract.Brace {
			switch tk.Val {
			case "[":
				braceCount++
				if braceCount == 1 {
					lastOpenBrace = i
				}
			case "]":
				braceCount--
			}
		}
		if braceCount > 0 {
			continue
		}
		if tk.Type == fract.Operator && tk.Val[len(tk.Val)-1] == '=' {
			setter = tk
			if lastOpenBrace == -1 {
				enumVal = p.processValuePart(valuePartInfo{mut: true, tokens: tokens[:i]})
				valTokens = tokens[i+1:]
				break
			}
			enumVal = p.processValuePart(valuePartInfo{mut: true, tokens: tokens[:lastOpenBrace]})
			valTokens = tokens[lastOpenBrace+1 : i-1]
			// Index value is empty?
			if len(valTokens) == 0 {
				fract.IPanic(setter, obj.SyntaxPanic, "Index is not given!")
			}
			selections = enumerableSelections(*enumVal, *p.processValTokens(valTokens), setter)
			valTokens = tokens[i+1:]
			break
		}
	}
	if len(valTokens) == 0 {
		fract.IPanicC(setter.File, setter.Line, setter.Column+len(setter.Val), obj.SyntaxPanic, "Value is not given!")
	}
	// Check const state.
	if enumVal.Const {
		fract.IPanic(setter, obj.SyntaxPanic, "Values is cannot changed of constant defines!")
	}
	val := *p.processValTokens(valTokens)
	if val.Data == nil {
		fract.IPanic(setter, obj.ValuePanic, "Invalid value!")
	}
	operator := obj.Token{Val: string(setter.Val[:len(setter.Val)-1])}
	if selections == nil {
		switch setter.Val {
		case "=": // =
			*enumVal = val
		default: // Other assignments.
			*enumVal = arithmeticProcess{
				operator: operator,
				left:     tokens,
				leftVal:  *enumVal,
				right:    []obj.Token{setter},
				rightVal: val,
			}.solve()
		}
		return
	}
	switch enumVal.Type {
	case oop.Map:
		m := enumVal.Data.(oop.MapModel)
		switch setter.Val {
		case "=":
			switch t := selections.(type) {
			case oop.ListModel:
				for _, key := range t.Elems {
					m.Map[key] = val
				}
			case oop.Val:
				m.Map[t] = val
			}
		default: // Other assignments.
			switch t := selections.(type) {
			case oop.ListModel:
				for _, key := range t.Elems {
					v, ok := m.Map[key]
					if !ok {
						m.Map[key] = val
						continue
					}
					m.Map[key] = arithmeticProcess{
						operator: operator,
						left:     tokens,
						leftVal:  v,
						right:    []obj.Token{setter},
						rightVal: val,
					}.solve()
				}
			case oop.Val:
				d, ok := m.Map[t]
				if !ok {
					m.Map[t] = val
					break
				}
				m.Map[t] = arithmeticProcess{
					operator: operator,
					left:     tokens,
					leftVal:  d,
					right:    []obj.Token{setter},
					rightVal: val,
				}.solve()
			}
		}
	case oop.List:
		for _, i := range selections.([]int) {
			switch setter.Val {
			case "=":
				enumVal.Data.(*oop.ListModel).Elems[i] = val
			default: // Other assignments.
				enumVal.Data.(*oop.ListModel).Elems[i] = arithmeticProcess{
					operator: operator,
					left:     tokens,
					leftVal:  enumVal.Data.(*oop.ListModel).Elems[i],
					right:    []obj.Token{setter},
					rightVal: val,
				}.solve()
			}
		}
	case oop.Str:
		for _, i := range selections.([]int) {
			switch setter.Val {
			case "=":
				if val.Type != oop.Str {
					fract.IPanic(setter, obj.ValuePanic, "Value type is not string!")
				} else if len(val.String()) > 1 {
					fract.IPanic(setter, obj.ValuePanic, "Value length is should be maximum one!")
				}
				bytes := []byte(enumVal.String())
				if val.Data == "" {
					bytes[i] = 0
				} else {
					bytes[i] = val.String()[0]
				}
				enumVal.Data = string(bytes)
			default: // Other assignments.
				val = arithmeticProcess{
					operator: operator,
					left:     tokens,
					leftVal:  oop.Val{Data: enumVal.Data.(string)[i], Type: oop.Int},
					right:    []obj.Token{setter},
					rightVal: val,
				}.solve()
				if val.Type != oop.Str {
					fract.IPanic(setter, obj.ValuePanic, "Value type is not string!")
				} else if len(val.String()) > 1 {
					fract.IPanic(setter, obj.ValuePanic, "Value length is should be maximum one!")
				}
				bytes := []byte(enumVal.String())
				if val.Data == "" {
					bytes[i] = 0
				} else {
					bytes[i] = val.String()[0]
				}
				enumVal.Data = string(bytes)
			}
		}
	}
}
