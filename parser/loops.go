package parser

import (
	"fmt"

	"github.com/fract-lang/fract/oop"
	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
)

// Loop.
type loop struct {
	a    oop.Val
	b    oop.Val
	enum oop.Val
	end  bool
}

func (l *loop) run(b func()) {
	switch l.enum.T {
	case oop.Array:
		l.a.T = oop.Int
		for i, e := range l.enum.D.(oop.ArrayModel) {
			l.a.D = fmt.Sprint(i)
			l.b = e
			b()
			if l.end {
				break
			}
		}
	case oop.Str:
		l.a.T = oop.Int
		l.b.T = oop.Str
		for i, e := range l.enum.D.(string) {
			l.a.D = fmt.Sprint(i)
			l.b.D = string(e)
			b()
			if l.end {
				break
			}
		}
	case oop.Map:
		for k, v := range l.enum.D.(oop.MapModel) {
			l.a = k
			l.b = v
			b()
			if l.end {
				break
			}
		}
	}
}

// Returns kwstate's return format.
func prockws(kws uint8) uint8 {
	if kws != fract.FUNCReturn {
		return fract.None
	}
	return kws
}

// Process loops and returns keyword state.
func (p *Parser) procLoop(tks []obj.Token) uint8 {
	bi := findBlock(tks)
	btks, tks := p.getBlock(tks[bi:]), tks[1:bi]
	flen := len(p.defs.Funcs)
	ilen := len(p.packages)
	brk := false
	kws := fract.None
	ptks := p.Tks
	pi := p.i
	//*************
	//    WHILE
	//*************
	if len(tks) == 0 || len(tks) >= 1 {
		if len(tks) == 0 || len(tks) == 1 || len(tks) >= 1 && tks[1].T != fract.In && tks[1].T != fract.Comma {
			vlen := len(p.defs.Vars)
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
				p.defs.Vars = p.defs.Vars[:vlen]
				// Remove temporary functions.
				p.defs.Funcs = p.defs.Funcs[:flen]
				// Remove temporary imports.
				p.packages = p.packages[:ilen]
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
			p.defs.Vars = p.defs.Vars[:vlen]
			// Remove temporary functions.
			p.defs.Funcs = p.defs.Funcs[:flen]
			// Remove temporary imports.
			p.packages = p.packages[:ilen]
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
	if nametk.V != "_" {
		if !validName(nametk.V) {
			fract.IPanic(nametk, obj.NamePanic, "Invalid name!")
		}
		if ln := p.definedName(nametk.V); ln != -1 {
			fract.IPanic(nametk, obj.NamePanic, "\""+nametk.V+"\" is already defined at line: "+fmt.Sprint(ln))
		}
	} else {
		nametk.V = ""
	}
	// Element name?
	ename := ""
	if tks[1].T == fract.Comma {
		if len(tks) < 3 || tks[2].T != fract.Name {
			fract.IPanic(tks[1], obj.SyntaxPanic, "Element name is not defined!")
		}
		if tks[2].V != "_" {
			ename = tks[2].V
			if !validName(ename) {
				fract.IPanic(tks[2], obj.NamePanic, "Invalid name!")
			}
			if ln := p.definedName(tks[2].V); ln != -1 {
				fract.IPanic(tks[2], obj.NamePanic, "\""+ename+"\" is already defined at line: "+fmt.Sprint(ln))
			}
		}
		if len(tks)-3 == 0 {
			tks[2].Col += len(tks[2].V)
			fract.IPanic(tks[2], obj.SyntaxPanic, "Value is not given!")
		}
		tks = tks[2:]
	}
	if len(tks) < 3 {
		fract.IPanic(tks[1], obj.SyntaxPanic, "Value is not given!")
	} else if t := tks[1]; t.T != fract.In && (t.T != fract.Operator || t.V != ":=") {
		fract.IPanic(tks[1], obj.SyntaxPanic, "Invalid syntax!")
	}
	tks = tks[2:]
	v := *p.procValTks(tks)
	// Type is not array?
	if !v.IsEnum() {
		fract.IPanic(tks[0], obj.ValuePanic, "Foreach loop must defined enumerable value!")
	}
	p.defs.Vars = append(p.defs.Vars,
		&oop.Var{Name: nametk.V, V: oop.Val{D: "0", T: oop.Int}},
		&oop.Var{Name: ename},
	)
	vlen := len(p.defs.Vars)
	index := p.defs.Vars[vlen-2]
	element := p.defs.Vars[vlen-1]
	vars := p.defs.Vars
	// Interpret block.
	l := loop{enum: v}
	l.run(func() {
		index.V = l.a
		element.V = l.b
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
		p.defs.Vars = vars
		// Remove temporary functions.
		p.defs.Funcs = p.defs.Funcs[:flen]
		// Remove temporary imports.
		p.packages = p.packages[:ilen]
		l.end = brk
	})
	p.Tks = ptks
	p.i = pi
	// Remove loop variables.
	index = nil
	element = nil
	p.defs.Vars = vars[:len(vars)-2]
	return prockws(kws)
}
