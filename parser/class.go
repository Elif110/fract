package parser

import (
	"fmt"

	"github.com/fract-lang/fract/oop"
	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
)

// buildClass from tokens.
func (p *Parser) buildClass(name string, tokens []obj.Token) *oop.Val {
	block := p.getBlock(tokens)
	class := oop.Class{Name: name, File: p.Lex.File}
	for _, tokens := range block {
		switch tokens[0].Type {
		case fract.Var:
			p.fvardec(&class.Defs, tokens)
		case fract.Fn:
			p.ffuncdec(&class.Defs, tokens)
			if f := class.Defs.Funcs[len(class.Defs.Funcs)-1]; f.Name == class.Name {
				if class.Constructor != nil {
					fract.IPanic(tokens[0], obj.NamePanic, "Constructor is already defined!")
				}
				class.Constructor = f
				class.Defs.Funcs = class.Defs.Funcs[:len(class.Defs.Funcs)-1]
			}
		default:
			fract.IPanic(tokens[0], obj.SyntaxPanic, "Invalid syntax!")
		}
	}
	if class.Constructor == nil { // Constructor is not given.
		class.Constructor = &oop.Fn{Name: class.Name + ".constructor", Src: p}
	}
	return &oop.Val{Data: class, Type: oop.ClassDef}
}

// Process class declaration.
func (p *Parser) classdec(tokens []obj.Token) {
	tokensLen := len(tokens)
	if tokensLen < 2 {
		fract.IPanic(tokens[0], obj.SyntaxPanic, "Invalid syntax!")
	}
	nameTk := tokens[1]
	if nameTk.Type != fract.Name {
		fract.IPanic(tokens[1], obj.SyntaxPanic, "Name is not valid!")
	}
	if line := p.defIndexByName(tokens[1].Val); line != -1 {
		fract.IPanic(tokens[1], obj.NamePanic, "\""+tokens[1].Val+"\" is already defined at line: "+fmt.Sprint(line))
	}
	classVal := *p.buildClass(nameTk.Val, tokens[2:])
	classVal.Const = true
	if p.funcTempVars != -1 {
		p.funcTempVars++
	}
	p.defs.Vars = append(p.defs.Vars, oop.Var{
		Name: nameTk.Val,
		Line: tokens[0].Line,
		Val:  classVal,
	})
}
