package parser

import (
	"fmt"

	"github.com/fract-lang/fract/oop"
	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
)

// buildClass from tokens.
func (p *Parser) buildClass(name string, tks []obj.Token) *oop.Val {
	blk := p.getBlock(tks)
	c := oop.Class{Name: name, L: p.L}
	for _, tks := range blk {
		switch tks[0].T {
		case fract.Var:
			p.fvardec(&c.Defs, tks)
		case fract.Func:
			p.ffuncdec(&c.Defs, tks)
			if f := c.Defs.Funcs[len(c.Defs.Funcs)-1]; f.Name == c.Name {
				if c.Constructor != nil {
					fract.IPanic(tks[0], obj.NamePanic, "Constructor is already defined!")
				}
				c.Constructor = f
				c.Defs.Funcs = c.Defs.Funcs[:len(c.Defs.Funcs)-1]
			}
		default:
			fract.IPanic(tks[0], obj.SyntaxPanic, "Invalid syntax!")
		}
	}
	if c.Constructor == nil { // Constructor is not given.
		c.Constructor = &oop.Func{Name: c.Name + ".constructor", Src: p}
	}
	return &oop.Val{D: c, T: oop.ClassDef}
}

// Process class declaration.
func (p *Parser) classdec(tks []obj.Token) {
	l := len(tks)
	if l < 2 {
		fract.IPanic(tks[0], obj.SyntaxPanic, "Invalid syntax!")
	}
	name := tks[1]
	if name.T != fract.Name {
		fract.IPanic(tks[1], obj.SyntaxPanic, "Name is not valid!")
	}
	if ln := p.definedName(tks[1].V); ln != -1 {
		fract.IPanic(tks[1], obj.NamePanic, "\""+tks[1].V+"\" is already defined at line: "+fmt.Sprint(ln))
	}
	v := *p.buildClass(name.V, tks[2:])
	v.Const = true
	if p.funcTempVars != -1 {
		p.funcTempVars++
	}
	p.defs.Vars = append(p.defs.Vars, &oop.Var{
		Name: name.V,
		Ln:   tks[0].Ln,
		V:    v,
	})
}
