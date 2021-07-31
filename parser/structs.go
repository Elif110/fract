package parser

import (
	"fmt"

	"github.com/fract-lang/fract/oop"
	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
)

// buildStruct from tokens.
func (p *Parser) buildStruct(name string, tks []obj.Token) *oop.Val {
	blk := p.getBlock(tks)
	s := oop.Struct{L: p.L}
	s.Constructor = &oop.Fn{Name: s.Name + ".constructor", Src: p}
	for _, tks := range blk {
		var comma bool
		for _, tk := range tks {
			switch tk.T {
			case fract.Comma:
				if !comma {
					fract.IPanic(tk, obj.SyntaxPanic, "Invalid syntax!")
				}
				comma = false
			case fract.Name:
				if !validName(tk.V) {
					fract.IPanic(tk, obj.NamePanic, "Invalid name!")
				}
				if comma {
					fract.IPanic(tk, obj.SyntaxPanic, "Invalid syntax!")
				}
				for _, p := range s.Constructor.Params {
					if p.Name == tk.V {
						fract.IPanic(tk, obj.NamePanic, "Field is already defined: "+tk.V)
					}
				}
				s.Constructor.Params = append(s.Constructor.Params, oop.Param{Name: tk.V})
				comma = true
			default:
				fract.IPanic(tk, obj.SyntaxPanic, "Invalid syntax!")
			}
		}
	}
	return &oop.Val{D: s, T: oop.StructDef}
}

// Process struct declaration.
func (p *Parser) structdec(tks []obj.Token) {
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
	v := *p.buildStruct(name.V, tks[2:])
	v.Const = true
	if p.funcTempVars != -1 {
		p.funcTempVars++
	}
	p.defs.Vars = append(p.defs.Vars, oop.Var{
		Name: name.V,
		Ln:   tks[0].Ln,
		V:    v,
	})
}
