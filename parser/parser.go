package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/fract-lang/fract/functions"
	"github.com/fract-lang/fract/lex"
	"github.com/fract-lang/fract/oop"
	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
)

var (
	defers []*funcCall
)

// Parser of Fract.
type Parser struct {
	defs         oop.DefMap
	packages     []importInfo
	funcTempVars int // Count of function temporary variables.
	loopCount    int
	funcCount    int
	i            int
	retVal       *oop.Val // Pointer of last return oop.
	pkg          string   // Package name.

	L   *lex.Lex
	Tks [][]obj.Token // All Tokens of code file.
}

// New returns instance of parser related to file.
func New(fp string) *Parser {
	f, _ := os.Open(fp)
	bytes, _ := os.ReadFile(fp)
	sf := &obj.File{P: fp, F: f}
	sf.Lns = strings.Split(string(bytes), "\n")
	for i, ln := range sf.Lns {
		sf.Lns[i] = strings.TrimRightFunc(ln, unicode.IsSpace)
	}
	return &Parser{
		funcTempVars: -1,
		L:            &lex.Lex{F: sf, Ln: 1},
	}
}

// NewStdin returns new instance of parser from standard input.
func NewStdin() *Parser {
	return &Parser{
		funcTempVars: -1,
		L: &lex.Lex{
			F:  &obj.File{P: "<stdin>"},
			Ln: 1,
		},
	}
}

// ready interpreter to process.
func (p *Parser) ready() {
	/// Tokenize all lines.
	for !p.L.Fin {
		if ctks := p.L.Next(); ctks != nil {
			p.Tks = append(p.Tks, ctks)
		}
	}
	// Detect package.
	if len(p.Tks) == 0 {
		fract.Error(p.L.F, 1, 1, "Package is not defined!")
	}
	tks := p.Tks[0]
	if tks[0].T != fract.Package {
		tk := tks[0]
		fract.Error(p.L.F, tk.Ln, tk.Col, "Package is must be define at first line!")
	}
	if len(tks) < 2 {
		tks[0].Col += len(tks[0].V)
		fract.IPanic(tks[0], obj.SyntaxPanic, "Package name is not given!")
	}
	n := tks[1]
	if n.T != fract.Name || !validName(n.V) {
		fract.IPanic(n, obj.SyntaxPanic, "Invalid package name!")
	}
	p.pkg = n.V
	if len(tks) > 2 {
		fract.IPanic(tks[2], obj.SyntaxPanic, "Invalid syntax!")
	}
}

func (p *Parser) importPackage() {
	dir, _ := os.Getwd()
	if pdir := path.Dir(p.L.F.P); pdir != "." {
		dir = path.Join(dir, pdir)
	}
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	_, mainName := filepath.Split(p.L.F.P)
	for _, info := range infos {
		// Skip directories.
		if info.IsDir() || !strings.HasSuffix(info.Name(), fract.Extension) || info.Name() == mainName {
			continue
		}
		src := New(path.Join(dir, info.Name()))
		src.ready()
		if src.pkg != p.pkg {
			tk := src.Tks[0][0]
			fract.Error(src.L.F, tk.Ln, tk.Col, "Package is not same!")
		}
		src.AddBuiltInFuncs()
		bifl := len(src.defs.Funcs)
		src.loopCount = -1 //! Tag as import source.
		src.Import()
		p.defs.Funcs = append(p.defs.Funcs, src.defs.Funcs[bifl:]...)
		p.defs.Vars = append(p.defs.Vars, src.defs.Vars...)
		p.packages = append(p.packages, src.packages...)
	}
}

func (p *Parser) Interpret() {
	if p.L.F.P == "<stdin>" {
		// Interpret all lines.
		for p.i = 0; p.i < len(p.Tks); p.i++ {
			p.process(p.Tks[p.i])
		}
		goto end
	}
	// Lexer is finished.
	if p.L.Fin {
		return
	}
	p.ready()
	p.importPackage()
	// Interpret all lines.
	for p.i = 1; p.i < len(p.Tks); p.i++ {
		p.process(p.Tks[p.i])
	}
end:
	for i := len(defers) - 1; i >= 0; i-- {
		defers[i].Call()
	}
}

// Process pragma.
func (p *Parser) procPragma(tks []obj.Token) {
	if tks[1].T != fract.Name {
		fract.IPanic(tks[0], obj.SyntaxPanic, "Invalid pragma!")
	}
	switch tks[1].V {
	case "enofi":
		if p.loopCount == -1 {
			p.loopCount = 0
		}
	default:
		fract.IPanic(tks[1], obj.SyntaxPanic, "Invalid pragma!")
	}
}

// validName returns true if name is valid, returns false if not.
func validName(n string) bool { return n != "_" && n != "this" }

// Process enumerable selections for access to elements.
func selections(enum, val oop.Val, tk obj.Token) interface{} {
	if val.T != oop.List && val.IsEnum() {
		fract.IPanic(tk, obj.ValuePanic, "Element selector is can only be list or single value!")
	}
	if enum.T == oop.Map {
		if val.T == oop.List {
			return val.D.(*oop.ListModel)
		}
		return val
	}

	// Array, String.
	l := enum.Len()
	if val.T == oop.List {
		var i []int
		for _, d := range val.D.(*oop.ListModel).Elems {
			if d.T != oop.Int {
				fract.IPanic(tk, obj.ValuePanic, "Only integer values can used in index access!")
			}
			pos, err := strconv.Atoi(d.String())
			if err != nil {
				fract.IPanic(tk, obj.OutOfRangePanic, "Value out of range!")
			}
			pos = procIndex(l, pos)
			if pos == -1 {
				fract.IPanic(tk, obj.OutOfRangePanic, "Index is out of range!")
			}
			i = append(i, pos)
		}
		return i
	}
	if val.T != oop.Int {
		fract.IPanic(tk, obj.ValuePanic, "Only integer values can used in index access!")
	}
	pos, err := strconv.Atoi(val.String())
	if err != nil {
		fract.IPanic(tk, obj.OutOfRangePanic, "Value out of range!")
	}
	pos = procIndex(l, pos)
	if pos == -1 {
		fract.IPanic(tk, obj.OutOfRangePanic, "Index is out of range!")
	}
	return []int{pos}
}

// Find start index of block.
func findBlock(tks []obj.Token) int {
	bc := 0
	for i, t := range tks {
		switch t.V {
		case "[", "(":
			bc++
		case "]", ")":
			bc--
		case "{":
			if bc == 0 {
				return i
			}
		}
	}
	fract.IPanic(tks[0], obj.SyntaxPanic, "Invalid syntax!")
	return -1
}

// Get a block.
func (p *Parser) getBlock(tks []obj.Token) [][]obj.Token {
	if len(tks) == 0 {
		p.i++
		tks = p.Tks[p.i]
	}
	if tks[0].T != fract.Brace && tks[0].V != "{" {
		fract.IPanic(tks[0], obj.SyntaxPanic, "Invalid syntax!")
	}
	bc := 0
	for i, t := range tks {
		if t.T == fract.Brace {
			switch t.V {
			case "{":
				bc++
			case "}":
				bc--
			}
		}
		if bc == 0 {
			if i < len(tks)-1 {
				p.Tks = append(p.Tks[:p.i+1], append([][]obj.Token{tks[i+1:]}, p.Tks[p.i+1:]...)...)
			}
			tks = tks[1 : i+1]
			break
		}
	}
	var btks [][]obj.Token
	if len(tks) == 1 {
		return btks
	}
	ln := tks[0].Ln
	lst := 0
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
		if t.T == fract.StatementTerminator {
			btks = append(btks, tks[lst:j])
			lst = j + 1
			ln = t.Ln
			continue
		}
		if ln < t.Ln {
			btks = append(btks, tks[lst:j])
			lst = j
			ln = t.Ln
		}
	}
	if len(tks) != lst {
		btks = append(btks, tks[lst:len(tks)-1])
	}
	return btks
}

// TYPES
// 'f' -> Function.
// 'v' -> Variable.
// 'p' -> Package.
// Returns define by name.
func (p *Parser) defByName(n string) (int, rune) {
	pos, t := p.defs.DefByName(n)
	if pos != -1 {
		return pos, t
	}
	pos = p.packageIndexByName(n)
	if pos != -1 {
		return pos, 'p'
	}
	return -1, '-'
}

// Returns index of name is exist name, returns -1 if not.
func (p *Parser) definedName(n string) int {
	if n[0] == '-' { // Ignore minus.
		n = n[1:]
	}
	for _, f := range p.defs.Funcs {
		if f.Name == n {
			return f.Ln
		}
	}
	for _, v := range p.defs.Vars {
		if v.Name == n {
			return v.Ln
		}
	}
	for _, i := range p.packages {
		if i.name == n {
			return i.ln
		}
	}
	return -1
}

// packageIndexByName returns index of package by name.
func (p *Parser) packageIndexByName(name string) int {
	for i, imp := range p.packages {
		if imp.name == name {
			return i
		}
	}
	return -1
}

// Check arithmetic processes validity.
func arithmeticProcesses(tks []obj.Token) [][]obj.Token {
	var (
		procs [][]obj.Token
		part  []obj.Token
		opr   bool
		b     int
	)
	for i := 0; i < len(tks); i++ {
		switch t := tks[i]; t.T {
		case fract.Operator:
			if !opr {
				fract.IPanic(t, obj.SyntaxPanic, "Operator overflow!")
			}
			opr = false
			if b > 0 {
				part = append(part, t)
			} else {
				procs = append(procs, part)
				procs = append(procs, []obj.Token{t})
				part = []obj.Token{}
			}
		default:
			if t.T == fract.Brace {
				switch t.V {
				case "(", "[", "{":
					b++
				default:
					b--
				}
			}
			if b == 0 && t.T == fract.Comma {
				fract.IPanic(t, obj.SyntaxPanic, "Invalid syntax!")
			}
			if i > 0 {
				if lt := tks[i-1]; (lt.T == fract.Name || lt.T == fract.Value) && (t.T == fract.Name || t.T == fract.Value) {
					fract.IPanic(t, obj.SyntaxPanic, "Invalid syntax!")
				}
			}
			part = append(part, t)
			opr = t.T != fract.Comma && (t.T != fract.Brace || t.T == fract.Brace && t.V != "[" && t.V != "(" && t.V != "{") && i < len(tks)-1
		}
	}
	if len(part) != 0 {
		procs = append(procs, part)
	}
	return procs
}

// decomposeBrace returns range tokens and index of first parentheses
// and remove range tokens from original tokens.
func decomposeBrace(tks *[]obj.Token) []obj.Token {
	fst := -1
	for i, t := range *tks {
		if t.T == fract.Brace && t.V == "(" {
			fst = i
			break
		}
	}
	// Skip find close parentheses and result ready steps
	// if open parentheses is not found.
	if fst == -1 {
		return nil
	}
	// Find close parentheses.
	c, l := 1, 0
	for i := fst + 1; i < len(*tks); i++ {
		tk := (*tks)[i]
		if tk.T == fract.Brace {
			switch tk.V {
			case "(":
				c++
			case ")":
				c--
			}
			if c == 0 {
				break
			}
		}
		l++
	}
	rg := append([]obj.Token{}, (*tks)[fst+1:fst+1+l]...)
	// Remove range from original tokens.
	*tks = append((*tks)[:fst], (*tks)[fst+2+l:]...)
	return rg
}

// procIndex process array index by length.
func procIndex(len, i int) int {
	if i >= 0 {
		if i >= len {
			return -1
		}
		return i
	}
	i = len + i
	if i < 0 || i >= len {
		return -1
	}
	return i
}

// nextopr find index of priority operator and returns index of operator
// if found, returns -1 if not.
func nextopr(tks [][]obj.Token) int {
	high, mid, low := -1, -1, -1
	for i, tslc := range tks {
		switch tslc[0].V {
		case "<<", ">>":
			return i
		case "**":
			return i
		case "%":
			return i
		case "*", "/", "\\", "//", "\\\\":
			if high == -1 {
				high = i
			}
		case "+", "-":
			if low == -1 {
				low = i
			}
		case "&", "|":
			if mid == -1 {
				mid = i
			}
		}
	}
	if high != -1 {
		return high
	} else if mid != -1 {
		return mid
	} else if low != -1 {
		return low
	}
	return -1
}

// findConditionOpr return next condition operator.
func findConditionOpr(tks []obj.Token) (int, obj.Token) {
	bc := 0
	for i, t := range tks {
		if t.T == fract.Brace {
			switch t.V {
			case "{", "[", "(":
				bc++
			default:
				bc--
			}
		}
		if bc > 0 {
			continue
		}
		switch t.T {
		case fract.Operator:
			switch t.V {
			case "&&", "||", "==", "<>", ">", "<", "<=", ">=":
				return i, t
			}
		case fract.In:
			return i, t
		}
	}
	var tk obj.Token
	return -1, tk
}

// Find next or condition operator index and return if find, return -1 if not.
func nextConditionOpr(tks []obj.Token, pos int, opr string) int {
	bc := 0
	for ; pos < len(tks); pos++ {
		t := tks[pos]
		if t.T == fract.Brace {
			switch t.V {
			case "{", "[", "(":
				bc++
			default:
				bc--
			}
		}
		if bc > 0 {
			continue
		}
		if t.T == fract.Operator && t.V == opr {
			return pos
		}
	}
	return -1
}

// conditionalProcesses returns conditional expressions by operators.
func conditionalProcesses(tks []obj.Token, opr string) [][]obj.Token {
	var exps [][]obj.Token
	last := 0
	i := nextConditionOpr(tks, last, opr)
	for i != -1 {
		if i-last == 0 {
			fract.IPanic(tks[last], obj.SyntaxPanic, "Condition expression is cannot given!")
		}
		exps = append(exps, tks[last:i])
		last = i + 1
		i = nextConditionOpr(tks, last, opr) // Find next.
		if i == len(tks)-1 {
			fract.IPanic(tks[len(tks)-1], obj.SyntaxPanic, "Operator overflow!")
		}
	}
	if last != len(tks) {
		exps = append(exps, tks[last:])
	}
	return exps
}

//! Built-in functions should have a lowercase names.

// ApplyBuildInFunctions to parser source.
func (p *Parser) AddBuiltInFuncs() {
	p.defs.Funcs = append(p.defs.Funcs,
		&oop.Fn{
			Name:          "print",
			DefParamCount: 2,
			Src:           functions.Print,
			Params: []oop.Param{{
				Name:   "value",
				Params: true,
				Defval: oop.Val{D: "", T: oop.Str},
			}},
		}, &oop.Fn{
			Name:          "println",
			Src:           functions.Println,
			DefParamCount: 2,
			Params: []oop.Param{{
				Name:   "value",
				Params: true,
				Defval: oop.Val{D: oop.NewListModel(oop.Val{D: "", T: oop.Str}), T: oop.List},
			}},
		}, &oop.Fn{
			Name:          "input",
			Src:           functions.Input,
			DefParamCount: 1,
			Params: []oop.Param{{
				Name:   "message",
				Defval: oop.Val{D: "", T: oop.Str},
			}},
		}, &oop.Fn{
			Name:          "exit",
			DefParamCount: 1,
			Src:           functions.Exit,
			Params: []oop.Param{{
				Name:   "code",
				Defval: oop.Val{D: "0", T: oop.Int},
			}},
		}, &oop.Fn{
			Name:          "len",
			Src:           functions.Len,
			DefParamCount: 0,
			Params:        []oop.Param{{Name: "object"}},
		}, &oop.Fn{
			Name:          "range",
			DefParamCount: 1,
			Src:           functions.Range,
			Params: []oop.Param{
				{Name: "start"},
				{Name: "to"},
				{
					Name:   "step",
					Defval: oop.Val{D: "1", T: oop.Int},
				},
			},
		}, &oop.Fn{
			Name:          "calloc",
			Src:           functions.Calloc,
			DefParamCount: 0,
			Params:        []oop.Param{{Name: "size"}},
		}, &oop.Fn{
			Name:          "realloc",
			DefParamCount: 0,
			Src:           functions.Realloc,
			Params:        []oop.Param{{Name: "base"}, {Name: "size"}},
		}, &oop.Fn{
			Name:          "string",
			Src:           functions.String,
			DefParamCount: 1,
			Params: []oop.Param{
				{Name: "object"},
				{
					Name:   "type",
					Defval: oop.Val{D: "parse", T: oop.Str},
				},
			},
		}, &oop.Fn{
			Name:          "int",
			Src:           functions.Int,
			DefParamCount: 1,
			Params: []oop.Param{
				{Name: "object"},
				{
					Name:   "type",
					Defval: oop.Val{D: "parse", T: oop.Str},
				},
			},
		}, &oop.Fn{
			Name:          "float",
			Src:           functions.Float,
			DefParamCount: 0,
			Params:        []oop.Param{{Name: "object"}},
		}, &oop.Fn{
			Name:          "panic",
			Src:           functions.Panic,
			DefParamCount: 0,
			Params:        []oop.Param{{Name: "msg"}},
		}, &oop.Fn{
			Name:          "type",
			Src:           functions.Type,
			DefParamCount: 0,
			Params:        []oop.Param{{Name: "obj"}},
		},
	)
}

// procIf process if-else if-else returns keyword state.
func (p *Parser) procIf(tks []obj.Token) uint8 {
	bi := findBlock(tks)
	btks, ctks := p.getBlock(tks[bi:]), tks[1:bi]
	// Condition is empty?
	if len(ctks) == 0 {
		first := tks[0]
		fract.IPanicC(first.F, first.Ln, first.Col+len(first.V), obj.SyntaxPanic, "Condition is empty!")
	}
	s := p.procCondition(ctks)
	vlen := len(p.defs.Vars)
	flen := len(p.defs.Funcs)
	ilen := len(p.packages)
	kws := fract.NA
	for _, tks := range btks {
		// Condition is true?
		if s == "true" && kws == fract.NA {
			if kws = p.process(tks); kws != fract.NA {
				break
			}
		} else {
			break
		}
	}
rep:
	p.i++
	if p.i >= len(p.Tks) {
		p.i--
		goto end
	}
	tks = p.Tks[p.i]
	if tks[0].T != fract.Else {
		p.i--
		goto end
	}
	if len(tks) > 1 && tks[1].T == fract.If { // Else if.
		bi = findBlock(tks)
		btks, ctks = p.getBlock(tks[bi:]), tks[2:bi]
		// Condition is empty?
		if len(ctks) == 0 {
			first := tks[1]
			fract.IPanicC(first.F, first.Ln, first.Col+len(first.V), obj.SyntaxPanic, "Condition is empty!")
		}
		if s == "true" {
			goto rep
		}
		s = p.procCondition(ctks)
		for _, tks := range btks {
			// Condition is true?
			if s == "true" && kws == fract.NA {
				if kws = p.process(tks); kws != fract.NA {
					break
				}
			} else {
				break
			}
		}
		goto rep
	}
	btks = p.getBlock(tks[1:])
	if s == "true" {
		goto end
	}
	for _, tks := range btks {
		// Condition is true?
		if kws == fract.NA {
			if kws = p.process(tks); kws != fract.NA {
				break
			}
		}
	}
end:
	p.defs.Vars = p.defs.Vars[:vlen]
	p.defs.Funcs = p.defs.Funcs[:flen]
	p.packages = p.packages[:ilen]
	return kws
}

func checkPublic(f *obj.File, n obj.Token) {
	if f != nil {
		if f == n.F {
			return
		}
	}
	if !unicode.IsUpper(rune(n.V[0])) {
		fract.IPanic(n, obj.NamePanic, "Name is not defined: "+n.V)
	}
}

// procTryCatch process try-catch blocks and returns keyword state.
func (p *Parser) procTryCatch(tks []obj.Token) uint8 {
	fract.TryCount++
	var (
		vlen = len(p.defs.Vars)
		flen = len(p.defs.Funcs)
		ilen = len(p.packages)
		dlen = len(defers)
		kws  = fract.NA
	)
	b := &obj.Block{
		Try: func() {
			for _, tks := range p.getBlock(tks[1:]) {
				if kws = p.process(tks); kws != fract.NA {
					break
				}
			}
			if p.Tks[p.i+1][0].T == fract.Catch {
				p.i++
			}
			fract.TryCount--
			p.defs.Vars = p.defs.Vars[:vlen]
			p.defs.Funcs = p.defs.Funcs[:flen]
			p.packages = p.packages[:ilen]
			for index := len(defers) - 1; index >= dlen; index-- {
				defers[index].Call()
			}
			defers = defers[:dlen]
		},
		Catch: func(cp obj.Panic) {
			for index := len(defers) - 1; index >= dlen; index-- {
				defers[index].Call()
			}
			p.loopCount = 0
			fract.TryCount--
			p.defs.Vars = p.defs.Vars[:vlen]
			p.defs.Funcs = p.defs.Funcs[:flen]
			p.packages = p.packages[:ilen]
			defers = defers[:dlen]
			p.i++
			tks = p.Tks[p.i]
			if tks[0].T != fract.Catch {
				p.i--
				return
			}
			l := len(tks)
			if l < 2 {
				fract.IPanic(tks[0], obj.SyntaxPanic, "Invalid syntax!")
			}
			n := tks[1]
			if n.T == fract.Name {
				if ln := p.definedName(n.V); ln != -1 {
					fract.IPanic(n, obj.NamePanic, "\""+n.V+"\" is already defined at line: "+fmt.Sprint(ln))
				}
				p.defs.Vars = append(p.defs.Vars, &oop.Var{
					Name: n.V,
					Ln:   n.Ln,
					V:    oop.Val{D: cp.M, T: oop.Str},
				})
			}
			var blk [][]obj.Token
			if n.T == fract.Name {
				if l < 3 {
					fract.IPanic(tks[1], obj.SyntaxPanic, "Invalid syntax!")
				}
				blk = p.getBlock(tks[2:])
			} else {
				blk = p.getBlock(tks[1:])
			}
			for _, tks := range blk {
				if kws = p.process(tks); kws != fract.NA {
					break
				}
			}
			p.defs.Vars = p.defs.Vars[:vlen]
			p.defs.Funcs = p.defs.Funcs[:flen]
			for i := len(defers) - 1; i >= dlen; i-- {
				defers[i].Call()
			}
			defers = defers[:dlen]
		},
	}
	b.Do()
	b = nil
	return kws
}

// TODO: Add match-case.
//! A change added here(especially added a code block) must also be compatible with "imports.go" and

// process tokens and returns true if block end, returns false if not and returns keyword state.
func (p *Parser) process(tks []obj.Token) uint8 {
	//tks = append([]obj.Token{}, tks...)
	switch fst := tks[0]; fst.T {
	case fract.Value, fract.Brace, fract.Name:
		if fst.T == fract.Name {
			bc := 0
			for _, t := range tks {
				if t.T == fract.Brace {
					switch t.V {
					case " {", "[", "(":
						bc++
					default:
						bc--
					}
				}
				if bc > 0 {
					continue
				}
				if t.T == fract.Operator {
					switch t.V {
					case "=", "+=", "-=", "*=", "/=", "%=", "^=", "<<=", ">>=", "|=", "&=":
						p.varset(tks)
						return fract.NA
					case ":=":
						p.varsdec(tks)
						return fract.NA
					}
				}
			}
		}
		// Print value if live interpreting.
		if v := p.procValTks(tks); fract.InteractiveSh {
			if v.Print() {
				println()
			}
		}
	case fract.Var:
		p.vardec(tks)
	case fract.If:
		return p.procIf(tks)
	case fract.Loop:
		p.loopCount++
		state := p.procLoop(tks)
		p.loopCount--
		return state
	case fract.Break:
		if p.loopCount < 1 {
			fract.IPanic(fst, obj.SyntaxPanic, "Break keyword only used in loops!")
		}
		return fract.LOOPBreak
	case fract.Continue:
		if p.loopCount < 1 {
			fract.IPanic(fst, obj.SyntaxPanic, "Continue keyword only used in loops!")
		}
		return fract.LOOPContinue
	case fract.Ret:
		if p.funcCount < 1 {
			fract.IPanic(fst, obj.SyntaxPanic, "Return keyword only used in functions!")
		}
		if len(tks) > 1 {
			value := p.procValTks(tks[1:])
			p.retVal = value
		} else {
			p.retVal = nil
		}
		return fract.FUNCReturn
	case fract.Fn:
		p.funcdec(tks)
	case fract.Try:
		return p.procTryCatch(tks)
	case fract.Import:
		p.procImport(tks)
	case fract.Macro:
		p.procPragma(tks)
	case fract.Struct:
		p.structdec(tks)
	case fract.Class:
		p.classdec(tks)
	case fract.Defer, fract.Go:
		if l := len(tks); l < 2 {
			fract.IPanic(tks[0], obj.SyntaxPanic, "Function is not given!")
		} else if t := tks[l-1]; t.T != fract.Brace && t.V != ")" {
			fract.IPanicC(tks[0].F, tks[0].Ln, tks[0].Col+len(tks[0].V), obj.SyntaxPanic, "Invalid syntax!")
		}
		var vtks []obj.Token
		bc := 0
		for i := len(tks) - 1; i >= 0; i-- {
			t := tks[i]
			if t.T != fract.Brace {
				continue
			}
			switch t.V {
			case ")":
				bc++
			case "(":
				bc--
			}
			if bc > 0 {
				continue
			}
			vtks = tks[1:i]
			break
		}
		if len(vtks) == 0 && bc == 0 {
			fract.IPanic(tks[1], obj.SyntaxPanic, "Invalid syntax!")
		}
		// Function call.
		v := p.procValPart(valPartInfo{tks: vtks})
		if v.T != oop.Func {
			fract.IPanic(tks[len(vtks)], obj.ValuePanic, "Value is not function!")
		}
		if fst.T == fract.Defer {
			defers = append(defers, p.funcCallModel(v.D.(*oop.Fn), tks[len(vtks):]))
		} else {
			go p.funcCallModel(v.D.(*oop.Fn), tks[len(vtks):]).Call()
		}
	default:
		fract.IPanic(fst, obj.SyntaxPanic, "Invalid syntax!")
	}
	return fract.NA
}
