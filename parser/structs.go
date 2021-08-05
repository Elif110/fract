package parser

import (
	"fmt"

	"github.com/fract-lang/fract/oop"
	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
)

// buildStruct from tokens.
func (p *Parser) buildStruct(name string, tokens []obj.Token) *oop.Val {
	block := p.getBlock(tokens)
	s := oop.Struct{Lex: p.Lex}
	s.Constructor = &oop.Fn{Name: s.Name + ".constructor", Src: p}
	for _, tokens := range block {
		var comma bool
		for _, tk := range tokens {
			switch tk.Type {
			case fract.Comma:
				if !comma {
					fract.IPanic(tk, obj.SyntaxPanic, "Invalid syntax!")
				}
				comma = false
			case fract.Name:
				if !isValidName(tk.Val) {
					fract.IPanic(tk, obj.NamePanic, "Invalid name!")
				}
				if comma {
					fract.IPanic(tk, obj.SyntaxPanic, "Invalid syntax!")
				}
				for _, param := range s.Constructor.Params {
					if param.Name == tk.Val {
						fract.IPanic(tk, obj.NamePanic, "Field is already defined: "+tk.Val)
					}
				}
				s.Constructor.Params = append(s.Constructor.Params, oop.Param{Name: tk.Val})
				comma = true
			default:
				fract.IPanic(tk, obj.SyntaxPanic, "Invalid syntax!")
			}
		}
	}
	return &oop.Val{Data: s, Type: oop.StructDef}
}

// Process struct declaration.
func (p *Parser) structdec(tokens []obj.Token) {
	tokensLen := len(tokens)
	if tokensLen < 2 {
		fract.IPanic(tokens[0], obj.SyntaxPanic, "Invalid syntax!")
	}
	nameTk := tokens[1]
	if nameTk.Type != fract.Name {
		fract.IPanic(tokens[1], obj.SyntaxPanic, "Name is not valid!")
	}
	if ln := p.defIndexByName(tokens[1].Val); ln != -1 {
		fract.IPanic(tokens[1], obj.NamePanic, "\""+tokens[1].Val+"\" is already defined at line: "+fmt.Sprint(ln))
	}
	val := *p.buildStruct(nameTk.Val, tokens[2:])
	val.Const = true
	if p.funcTempVars != -1 {
		p.funcTempVars++
	}
	p.defs.Vars = append(p.defs.Vars, &oop.Var{
		Name: nameTk.Val,
		Line: tokens[0].Line,
		Val:  val,
	})
}
