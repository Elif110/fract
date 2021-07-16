package parser

import (
	"fmt"

	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
	"github.com/fract-lang/fract/pkg/value"
)

// prockws returns return value of kwstate.
func prockws(kws uint8) uint8 {
	if kws != fract.FUNCReturn {
		return fract.None
	}
	return kws
}

// procLoop process loops and returns keyword state.
func (p *Parser) procLoop(tks obj.Tokens) uint8 {
	bi := findBlock(tks)
	btks, tks := p.getBlock(tks[bi:]), tks[1:bi]
	flen := len(p.funcs)
	ilen := len(p.Imports)
	brk := false
	kws := fract.None
	ptks := p.Tks
	pi := p.i
	//*************
	//    WHILE
	//*************
	if len(tks) == 0 || len(tks) >= 1 {
		if len(tks) == 0 || len(tks) == 1 || len(tks) >= 1 && tks[1].T != fract.In && tks[1].T != fract.Comma {
			vlen := len(p.vars)
			// Infinity loop.
			if len(tks) == 0 {
			infinity:
				p.Tks = btks
				for p.i = 0; p.i < len(p.Tks); p.i++ {
					kws = p.process(p.Tks[p.i])
					if kws == fract.LOOPBreak || kws == fract.FUNCReturn { // Break loop or return.
						p.Tks = ptks
						p.i = pi
						return prockws(kws)
					} else if kws == fract.LOOPContinue { // Continue loop.
						break
					}
				}
				// Remove temporary variables.
				p.vars = p.vars[:vlen]
				// Remove temporary functions.
				p.funcs = p.funcs[:flen]
				// Remove temporary imports.
				p.Imports = p.Imports[:ilen]
				goto infinity
			}
		while:
			// Interpret/skip block.
			c := p.procCondition(tks)
			p.Tks = btks
			for p.i = 0; p.i < len(p.Tks); p.i++ {
				// Condition is true?
				if c == "true" {
					kws = p.process(p.Tks[p.i])
					if kws == fract.LOOPBreak || kws == fract.FUNCReturn { // Break loop or return.
						brk = true
						break
					} else if kws == fract.LOOPContinue { // Continue loop.
						break
					}
				} else {
					brk = true
					break
				}
			}
			// Remove temporary variables.
			p.vars = p.vars[:vlen]
			// Remove temporary functions.
			p.funcs = p.funcs[:flen]
			// Remove temporary imports.
			p.Imports = p.Imports[:ilen]
			c = p.procCondition(tks)
			if brk || c != "true" {
				p.Tks = ptks
				p.i = pi
				return prockws(kws)
			}
			goto while
		}
	}
	//*************
	//   FOREACH
	//*************
	nametk := tks[0]
	// Name is not name?
	if nametk.T != fract.Name {
		fract.IPanic(nametk, obj.SyntaxPanic, "This is not a valid name!")
	}
	if ln := p.definedName(nametk); ln != -1 {
		fract.IPanic(nametk, obj.NamePanic, "\""+nametk.V+"\" is already defined at line: "+fmt.Sprint(ln))
	}
	// Element name?
	ename := ""
	if tks[1].T == fract.Comma {
		if len(tks) < 3 || tks[2].T != fract.Name {
			fract.IPanic(tks[1], obj.SyntaxPanic, "Element name is not defined!")
		}
		if tks[2].V != "_" {
			ename = tks[2].V
			if ln := p.definedName(tks[2]); ln != -1 {
				fract.IPanic(tks[2], obj.NamePanic, "\""+ename+"\" is already defined at line: "+fmt.Sprint(ln))
			}
		}
		if len(tks)-3 == 0 {
			tks[2].Col += len(tks[2].V)
			fract.IPanic(tks[2], obj.SyntaxPanic, "Value is not given!")
		}
		tks = tks[2:]
	}
	if vtks, inTk := tks.Sub(2, len(tks)-2), tks[1]; vtks != nil {
		tks = *vtks
	} else {
		fract.IPanic(inTk, obj.SyntaxPanic, "Value is not given!")
	}
	v := p.procVal(tks)
	// Type is not array?
	if v.T != value.Array && v.T != value.Map && v.D[0].T != value.Str {
		fract.IPanic(tks[0], obj.ValuePanic, "Foreach loop must defined enumerable value!")
	}
	// TODO: Add Map.
	p.vars = append(p.vars,
		obj.Var{Name: nametk.V, V: value.Val{D: []value.Data{{D: "0", T: value.Int}}}},
		obj.Var{Name: ename, V: value.Val{}},
	)
	vlen := len(p.vars)
	index := &p.vars[vlen-2]
	element := &p.vars[vlen-1]
	vars := p.vars
	if index.Name == "_" {
		index.Name = ""
	}
	var length int
	if v.T == value.Array {
		length = len(v.D)
	} else {
		length = len(v.D[0].String())
	}
	if element.Name != "" {
		if v.T == value.Array {
			element.V.D = []value.Data{v.D[0]}
		} else {
			element.V.D = []value.Data{{D: string(v.D[0].String()[0]), T: value.Str}}
		}
	}
	// Interpret block.
	for j := 1; ; j++ {
		p.Tks = btks
		for p.i = 0; p.i < len(p.Tks); p.i++ {
			kws = p.process(p.Tks[p.i])
			if kws == fract.LOOPBreak || kws == fract.FUNCReturn { // Break loop or return.
				brk = true
				break
			} else if kws == fract.LOOPContinue { // Continue loop.
				break
			}
		}
		// Remove temporary variables.
		p.vars = vars
		// Remove temporary functions.
		p.funcs = p.funcs[:flen]
		// Remove temporary imports.
		p.Imports = p.Imports[:ilen]
		if brk || j >= length {
			break
		}
		if index.Name != "" {
			index.V.D = []value.Data{{D: fmt.Sprint(j), T: value.Int}}
		}
		if element.Name != "" {
			if v.T == value.Array {
				element.V.D = []value.Data{v.D[j]}
			} else {
				element.V.D = []value.Data{{D: string(v.D[0].String()[j]), T: value.Str}}
			}
		}
	}
	p.Tks = ptks
	p.i = pi
	// Remove loop variables.
	p.vars = vars[:len(vars)-2]
	return prockws(kws)
}
