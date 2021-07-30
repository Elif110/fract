package parser

import (
	"fmt"

	"github.com/fract-lang/fract/oop"
	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
)

// Metadata of variable declaration.
type varinfo struct {
	sdec     bool
	constant bool
	mut      bool
}

// Append variable to source.
func (p *Parser) varadd(dm *oop.DefMap, md varinfo, tks []obj.Token) {
	name := tks[0]
	if !validName(name.V) {
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

	tksLen := len(tks)
	// Setter is not defined?
	if tksLen < 2 {
		fract.IPanicC(name.F, name.Ln, name.Col+len(name.V), obj.SyntaxPanic, "Setter is not found!")
	}
	setter := tks[1]
	// Setter is not a setter operator?
	if setter.T != fract.Operator || (setter.V != "=" && !md.sdec || setter.V != ":=" && md.sdec) {
		fract.IPanic(setter, obj.SyntaxPanic, "Invalid setter operator: "+setter.V)
	}
	// Value is not defined?
	if tksLen < 3 {
		fract.IPanicC(setter.F, setter.Ln, setter.Col+len(setter.V), obj.SyntaxPanic, "Value is not given!")
	}
	v := *p.procValTks(tks[2:])
	if v.D == nil {
		fract.IPanic(tks[2], obj.ValuePanic, "Invalid value!")
	}
	if p.funcTempVars != -1 {
		p.funcTempVars++
	}
	v.Mut = md.mut
	v.Const = md.constant
	dm.Vars = append(dm.Vars, &oop.Var{
		Name: name.V,
		V:    v,
		Ln:   name.Ln,
	})
}

// Process variable declaration to defmap.
func (p *Parser) fvardec(dm *oop.DefMap, tks []obj.Token) {
	// Name is not defined?
	if len(tks) < 2 {
		first := tks[0]
		fract.IPanicC(first.F, first.Ln, first.Col+len(first.V), obj.SyntaxPanic, "Name is not given!")
	}
	md := varinfo{
		constant: tks[0].V == "const",
		mut:      tks[0].V == "mut",
	}
	pre := tks[1]
	if pre.T == fract.Name {
		p.varadd(dm, md, tks[1:])
	} else if pre.T == fract.Brace && pre.V == "(" {
		tks = tks[2 : len(tks)-1]
		lst := 0
		ln := tks[0].Ln
		bc := 0
		for j, t := range tks {
			if t.T == fract.Brace {
				switch t.V {
				case "{", "[", "(":
					bc++
				default:
					bc--
					ln = t.Ln
				}
			}
			if bc > 0 {
				continue
			}
			if ln < t.Ln {
				p.varadd(dm, md, tks[lst:j])
				lst = j
				ln = t.Ln
			}
		}
		if len(tks) != lst {
			p.varadd(dm, md, tks[lst:])
		}
	} else {
		fract.IPanic(pre, obj.SyntaxPanic, "Invalid syntax!")
	}
}

// Process variable declaration to parser.
func (p *Parser) vardec(tks []obj.Token) { p.fvardec(&p.defs, tks) }

// Process short variable declaration.
func (p *Parser) varsdec(tks []obj.Token) {
	// Name is not defined?
	if len(tks) < 2 {
		first := tks[0]
		fract.IPanicC(first.F, first.Ln, first.Col+len(first.V), obj.SyntaxPanic, "Name is not given!")
	}
	if tks[0].T != fract.Name {
		fract.IPanic(tks[0], obj.SyntaxPanic, "Invalid syntax!")
	}
	var md varinfo
	md.sdec = true
	p.varadd(&p.defs, md, tks)
}

// Process variable set statement.
func (p *Parser) varset(tks []obj.Token) {
	var (
		v      *oop.Val
		s      interface{}
		vtks   []obj.Token
		setter obj.Token
	)
	bc := 0
	lbc := -1
	for i, tk := range tks {
		if tk.T == fract.Brace {
			switch tk.V {
			case "[":
				bc++
				if bc == 1 {
					lbc = i
				}
			case "]":
				bc--
			}
		}
		if bc > 0 {
			continue
		}
		if tk.T == fract.Operator && tk.V[len(tk.V)-1] == '=' {
			setter = tk
			if lbc == -1 {
				v = p.procValPart(valPartInfo{mut: true, tks: tks[:i]})
				vtks = tks[i+1:]
				break
			}
			v = p.procValPart(valPartInfo{mut: true, tks: tks[:lbc]})
			vtks = tks[lbc+1 : i-1]
			// Index value is empty?
			if len(vtks) == 0 {
				fract.IPanic(setter, obj.SyntaxPanic, "Index is not given!")
			}
			s = selections(*v, *p.procValTks(vtks), setter)
			vtks = tks[i+1:]
			break
		}
	}
	if len(vtks) == 0 {
		fract.IPanicC(setter.F, setter.Ln, setter.Col+len(setter.V), obj.SyntaxPanic, "Value is not given!")
	}
	// Check const state.
	if v.Const {
		fract.IPanic(setter, obj.SyntaxPanic, "Values is cannot changed of constant defines!")
	}
	val := *p.procValTks(vtks)
	if val.D == nil {
		fract.IPanic(setter, obj.ValuePanic, "Invalid value!")
	}
	opr := obj.Token{V: string(setter.V[:len(setter.V)-1])}
	if s == nil {
		switch setter.V {
		case "=": // =
			*v = val
		default: // Other assignments.
			*v = solveProc(process{
				opr: opr,
				f:   tks,
				fv:  *v,
				s:   []obj.Token{setter},
				sv:  val,
			})
		}
		return
	}
	switch v.T {
	case oop.Map:
		m := v.D.(oop.MapModel)
		switch setter.V {
		case "=":
			switch t := s.(type) {
			case oop.ArrayModel:
				for _, k := range t {
					m.M[k] = val
				}
			case oop.Val:
				m.M[t] = val
			}
		default: // Other assignments.
			switch t := s.(type) {
			case oop.ArrayModel:
				for _, s := range t {
					d, ok := m.M[s]
					if !ok {
						m.M[s] = val
						continue
					}
					m.M[s] = solveProc(process{
						opr: opr,
						f:   tks,
						fv:  d,
						s:   []obj.Token{setter},
						sv:  val,
					})
				}
			case oop.Val:
				d, ok := m.M[t]
				if !ok {
					m.M[t] = val
					break
				}
				m.M[t] = solveProc(process{
					opr: opr,
					f:   tks,
					fv:  d,
					s:   []obj.Token{setter},
					sv:  val,
				})
			}
		}
	case oop.Array:
		for _, pos := range s.([]int) {
			switch setter.V {
			case "=":
				v.D.(oop.ArrayModel)[pos] = val
			default: // Other assignments.
				v.D.(oop.ArrayModel)[pos] = solveProc(process{
					opr: opr,
					f:   tks,
					fv:  v.D.(oop.ArrayModel)[pos],
					s:   []obj.Token{setter},
					sv:  val,
				})
			}
		}
	case oop.Str:
		for _, pos := range s.([]int) {
			switch setter.V {
			case "=":
				if val.T != oop.Str {
					fract.IPanic(setter, obj.ValuePanic, "Value type is not string!")
				} else if len(val.String()) > 1 {
					fract.IPanic(setter, obj.ValuePanic, "Value length is should be maximum one!")
				}
				bytes := []byte(v.String())
				if val.D == "" {
					bytes[pos] = 0
				} else {
					bytes[pos] = val.String()[0]
				}
				v.D = string(bytes)
			default: // Other assignments.
				val = solveProc(process{
					opr: opr,
					f:   tks,
					fv:  oop.Val{D: v.D.(string)[pos], T: oop.Int},
					s:   []obj.Token{setter},
					sv:  val,
				})
				if val.T != oop.Str {
					fract.IPanic(setter, obj.ValuePanic, "Value type is not string!")
				} else if len(val.String()) > 1 {
					fract.IPanic(setter, obj.ValuePanic, "Value length is should be maximum one!")
				}
				bytes := []byte(v.String())
				if val.D == "" {
					bytes[pos] = 0
				} else {
					bytes[pos] = val.String()[0]
				}
				v.D = string(bytes)
			}
		}
	}
}
