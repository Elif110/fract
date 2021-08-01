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
	index        int
	returnVal    *oop.Val // Pointer of last return oop.
	packageName  string   // Package name.

	Lex    *lex.Lex
	Tokens [][]obj.Token // All Tokens of code file.
}

// New returns instance of parser related to file.
func New(fp string) *Parser {
	file, _ := os.Open(fp)
	bytes, _ := os.ReadFile(fp)
	fileObj := &obj.File{Path: fp, File: file}
	fileObj.Lines = strings.Split(string(bytes), "\n")
	for i, line := range fileObj.Lines {
		fileObj.Lines[i] = strings.TrimRightFunc(line, unicode.IsSpace)
	}
	return &Parser{
		funcTempVars: -1,
		Lex:          &lex.Lex{File: fileObj, Line: 1},
	}
}

// NewStdin returns new instance of parser from standard input.
func NewStdin() *Parser {
	return &Parser{
		funcTempVars: -1,
		Lex: &lex.Lex{
			File: &obj.File{Path: "<stdin>"},
			Line: 1,
		},
	}
}

// ready interpreter to process.
func (p *Parser) ready() {
	/// Tokenize all lines.
	for !p.Lex.Finished {
		if ctks := p.Lex.Next(); ctks != nil {
			p.Tokens = append(p.Tokens, ctks)
		}
	}
	// Detect package.
	if len(p.Tokens) == 0 {
		fract.Error(p.Lex.File, 1, 1, "Package is not defined!")
	}
	tokens := p.Tokens[0]
	if tokens[0].Type != fract.Package {
		tk := tokens[0]
		fract.Error(p.Lex.File, tk.Line, tk.Column, "Package is must be define at first line!")
	}
	if len(tokens) < 2 {
		tokens[0].Column += len(tokens[0].Val)
		fract.IPanic(tokens[0], obj.SyntaxPanic, "Package name is not given!")
	}
	nameTk := tokens[1]
	if nameTk.Type != fract.Name || !isValidName(nameTk.Val) {
		fract.IPanic(nameTk, obj.SyntaxPanic, "Invalid package name!")
	}
	p.packageName = nameTk.Val
	if len(tokens) > 2 {
		fract.IPanic(tokens[2], obj.SyntaxPanic, "Invalid syntax!")
	}
}

func (p *Parser) importPackage() {
	dir, _ := os.Getwd()
	if srcPathDir := path.Dir(p.Lex.File.Path); srcPathDir != "." {
		dir = path.Join(dir, srcPathDir)
	}
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	_, mainName := filepath.Split(p.Lex.File.Path)
	for _, info := range infos {
		// Skip directories.
		if info.IsDir() || !strings.HasSuffix(info.Name(), fract.Extension) || info.Name() == mainName {
			continue
		}
		src := New(path.Join(dir, info.Name()))
		src.ready()
		if src.packageName != p.packageName {
			tk := src.Tokens[0][0]
			fract.Error(src.Lex.File, tk.Line, tk.Column, "Package is not same!")
		}
		src.AddBuiltInFuncs()
		builtinFuncLen := len(src.defs.Funcs)
		src.loopCount = -1 //! Tag as import source.
		src.Import()
		p.defs.Funcs = append(p.defs.Funcs, src.defs.Funcs[builtinFuncLen:]...)
		p.defs.Vars = append(p.defs.Vars, src.defs.Vars...)
		p.packages = append(p.packages, src.packages...)
	}
}

func (p *Parser) Interpret() {
	if p.Lex.File.Path == "<stdin>" {
		// Interpret all lines.
		for p.index = 0; p.index < len(p.Tokens); p.index++ {
			p.processExpression(p.Tokens[p.index])
		}
		goto end
	}
	// Lexer is finished.
	if p.Lex.Finished {
		return
	}
	p.ready()
	p.importPackage()
	// Interpret all lines.
	for p.index = 1; p.index < len(p.Tokens); p.index++ {
		p.processExpression(p.Tokens[p.index])
	}
end:
	for i := len(defers) - 1; i >= 0; i-- {
		defers[i].Call()
	}
}

func (p *Parser) processPragma(tks []obj.Token) {
	if tks[1].Type != fract.Name {
		fract.IPanic(tks[0], obj.SyntaxPanic, "Invalid pragma!")
	}
	switch tks[1].Val {
	case "enofi":
		if p.loopCount == -1 {
			p.loopCount = 0
		}
	default:
		fract.IPanic(tks[1], obj.SyntaxPanic, "Invalid pragma!")
	}
}

// isValidName returns true if name is valid, returns false if not.
func isValidName(name string) bool { return name != "_" && name != "this" }

// enumerableSelections process enumerable enumerableSelections for access to elements.
func enumerableSelections(enum, selectVal oop.Val, tk obj.Token) interface{} {
	if enum.Type == oop.Map {
		if selectVal.Type == oop.List {
			return selectVal.Data.(*oop.ListModel)
		}
		return selectVal
	}

	if selectVal.Type != oop.List && selectVal.IsEnum() {
		fract.IPanic(tk, obj.ValuePanic, "Element selector is can only be list or single value!")
	}
	// List, String.
	enumLen := enum.Len()
	if selectVal.Type == oop.List {
		var i []int
		for _, d := range selectVal.Data.(*oop.ListModel).Elems {
			if d.Type != oop.Int {
				fract.IPanic(tk, obj.ValuePanic, "Only integer values can used in index access!")
			}
			pos, err := strconv.Atoi(d.String())
			if err != nil {
				fract.IPanic(tk, obj.OutOfRangePanic, "Value out of range!")
			}
			pos = processIndex(enumLen, pos)
			if pos == -1 {
				fract.IPanic(tk, obj.OutOfRangePanic, "Index is out of range!")
			}
			i = append(i, pos)
		}
		return i
	}
	if selectVal.Type != oop.Int {
		fract.IPanic(tk, obj.ValuePanic, "Only integer values can used in index access!")
	}
	pos, err := strconv.Atoi(selectVal.String())
	if err != nil {
		fract.IPanic(tk, obj.OutOfRangePanic, "Value out of range!")
	}
	pos = processIndex(enumLen, pos)
	if pos == -1 {
		fract.IPanic(tk, obj.OutOfRangePanic, "Index is out of range!")
	}
	return []int{pos}
}

// findBlock returns start index of block if found, returns -1 if not.
func findBlock(tokens []obj.Token) int {
	braceCount := 0
	for i, tk := range tokens {
		switch tk.Val {
		case "[", "(":
			braceCount++
		case "]", ")":
			braceCount--
		case "{":
			if braceCount == 0 {
				return i
			}
		}
	}
	fract.IPanic(tokens[0], obj.SyntaxPanic, "Block is not given!")
	return -1
}

func (p *Parser) getBlock(tokens []obj.Token) [][]obj.Token {
	if len(tokens) == 0 {
		p.index++
		tokens = p.Tokens[p.index]
	}
	if tokens[0].Type != fract.Brace || tokens[0].Val != "{" {
		fract.IPanic(tokens[0], obj.SyntaxPanic, "Invalid syntax!")
	}
	braceCount := 0
	for i, tk := range tokens {
		if tk.Type == fract.Brace {
			switch tk.Val {
			case "{":
				braceCount++
			case "}":
				braceCount--
			}
		}
		if braceCount == 0 {
			if i < len(tokens)-1 {
				p.Tokens = append(p.Tokens[:p.index+1], append([][]obj.Token{tokens[i+1:]}, p.Tokens[p.index+1:]...)...)
			}
			tokens = tokens[1 : i+1]
			break
		}
	}
	var blockTokens [][]obj.Token
	if len(tokens) == 1 {
		return blockTokens
	}
	line := tokens[0].Line
	lastIndex := 0
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
		if tk.Type == fract.StatementTerminator {
			blockTokens = append(blockTokens, tokens[lastIndex:j])
			lastIndex = j + 1
			line = tk.Line
			continue
		}
		if line < tk.Line {
			blockTokens = append(blockTokens, tokens[lastIndex:j])
			lastIndex = j
			line = tk.Line
		}
	}
	if len(tokens) != lastIndex {
		blockTokens = append(blockTokens, tokens[lastIndex:len(tokens)-1])
	}
	return blockTokens
}

// TYPES
// 'f' -> Function.
// 'v' -> Variable.
// 'p' -> Package.
// Returns define by name.
func (p *Parser) defByName(n string) (int, rune) {
	pos, typ := p.defs.DefByName(n)
	if pos != -1 {
		return pos, typ
	}
	pos = p.packageIndexByName(n)
	if pos != -1 {
		return pos, 'p'
	}
	return -1, '-'
}

// defIndexByName returns index of name is exist name, returns -1 if not.
func (p *Parser) defIndexByName(name string) int {
	if name[0] == '-' { // Ignore minus.
		name = name[1:]
	}
	for _, f := range p.defs.Funcs {
		if f.Name == name {
			return f.Line
		}
	}
	for _, v := range p.defs.Vars {
		if v.Name == name {
			return v.Line
		}
	}
	for _, i := range p.packages {
		if i.name == name {
			return i.line
		}
	}
	return -1
}

func (p *Parser) packageIndexByName(name string) int {
	for i, imp := range p.packages {
		if imp.name == name {
			return i
		}
	}
	return -1
}

func arithmeticProcesses(tokens []obj.Token) [][]obj.Token {
	var (
		processes  [][]obj.Token
		part       []obj.Token
		operator   bool
		braceCount int
	)
	for i := 0; i < len(tokens); i++ {
		switch tk := tokens[i]; tk.Type {
		case fract.Operator:
			if !operator {
				fract.IPanic(tk, obj.SyntaxPanic, "Operator overflow!")
			}
			operator = false
			if braceCount > 0 {
				part = append(part, tk)
			} else {
				processes = append(processes, part)
				processes = append(processes, []obj.Token{tk})
				part = []obj.Token{}
			}
		default:
			if tk.Type == fract.Brace {
				switch tk.Val {
				case "(", "[", "{":
					braceCount++
				default:
					braceCount--
				}
			}
			if braceCount == 0 && tk.Type == fract.Comma {
				fract.IPanic(tk, obj.SyntaxPanic, "Invalid syntax!")
			}
			if i > 0 {
				if lt := tokens[i-1]; (lt.Type == fract.Name || lt.Type == fract.Value) && (tk.Type == fract.Name || tk.Type == fract.Value) {
					fract.IPanic(tk, obj.SyntaxPanic, "Invalid syntax!")
				}
			}
			part = append(part, tk)
			operator = tk.Type != fract.Comma && (tk.Type != fract.Brace || tk.Type == fract.Brace && tk.Val != "[" && tk.Val != "(" && tk.Val != "{") && i < len(tokens)-1
		}
	}
	if len(part) != 0 {
		processes = append(processes, part)
	}
	return processes
}

// decomposeBrace returns range tokens and index of first parentheses
// and remove range tokens from original tokens.
func decomposeBrace(tokens *[]obj.Token) []obj.Token {
	first := -1
	for i, tk := range *tokens {
		if tk.Type == fract.Brace && tk.Val == "(" {
			first = i
			break
		}
	}
	// Skip find close parentheses and result ready steps
	// if open parentheses is not found.
	if first == -1 {
		return nil
	}
	// Find close parentheses.
	braceCount, length := 1, 0
	for i := first + 1; i < len(*tokens); i++ {
		tk := (*tokens)[i]
		if tk.Type == fract.Brace {
			switch tk.Val {
			case "(":
				braceCount++
			case ")":
				braceCount--
			}
			if braceCount == 0 {
				break
			}
		}
		length++
	}
	rangeTokens := append([]obj.Token{}, (*tokens)[first+1:first+1+length]...)
	// Remove range from original tokens.
	*tokens = append((*tokens)[:first], (*tokens)[first+2+length:]...)
	return rangeTokens
}

// processIndex is process index by length.
func processIndex(length, index int) int {
	if index >= 0 {
		if index >= length {
			return -1
		}
		return index
	}
	index = length + index
	if index < 0 || index >= length {
		return -1
	}
	return index
}

// nextOperator find index of priority operator and returns index of operator
// if found, returns -1 if not.
func nextOperator(tokens [][]obj.Token) int {
	high, mid, low := -1, -1, -1
	for i, part := range tokens {
		switch part[0].Val {
		case "<<", ">>":
			return i
		case "**":
			return i
		case "%":
			return i
		case "*", "/", "\\":
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

// findConditionOperator return next condition operator.
func findConditionOperator(tokens []obj.Token) (int, obj.Token) {
	braceCount := 0
	for i, tk := range tokens {
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
		switch tk.Type {
		case fract.Operator:
			switch tk.Val {
			case "&&", "||", "==", "<>", ">", "<", "<=", ">=":
				return i, tk
			}
		case fract.In:
			return i, tk
		}
	}
	var tk obj.Token
	return -1, tk
}

// Find next or condition operator index and return if find, return -1 if not.
func nextConditionOperator(tokens []obj.Token, pos int, operator string) int {
	braceCount := 0
	for ; pos < len(tokens); pos++ {
		tk := tokens[pos]
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
		if tk.Type == fract.Operator && tk.Val == operator {
			return pos
		}
	}
	return -1
}

// conditionalProcesses returns conditional expressions by operators.
func conditionalProcesses(tokens []obj.Token, operator string) [][]obj.Token {
	var expressions [][]obj.Token
	last := 0
	index := nextConditionOperator(tokens, last, operator)
	for index != -1 {
		if index-last == 0 {
			fract.IPanic(tokens[last], obj.SyntaxPanic, "Condition expression is cannot given!")
		}
		expressions = append(expressions, tokens[last:index])
		last = index + 1
		index = nextConditionOperator(tokens, last, operator) // Find next.
		if index == len(tokens)-1 {
			fract.IPanic(tokens[len(tokens)-1], obj.SyntaxPanic, "Operator overflow!")
		}
	}
	if last != len(tokens) {
		expressions = append(expressions, tokens[last:])
	}
	return expressions
}

//! Built-in functions should have a lowercase names.

func (p *Parser) AddBuiltInFuncs() {
	p.defs.Funcs = append(p.defs.Funcs,
		&oop.Fn{
			Name:              "print",
			DefaultParamCount: 2,
			Src:               functions.Print,
			Params: []oop.Param{{
				Name:       "value",
				Params:     true,
				DefaultVal: oop.Val{Data: "", Type: oop.String},
			}},
		}, &oop.Fn{
			Name:              "println",
			Src:               functions.Println,
			DefaultParamCount: 2,
			Params: []oop.Param{{
				Name:       "value",
				Params:     true,
				DefaultVal: oop.Val{Data: oop.NewListModel(oop.Val{Data: "", Type: oop.String}), Type: oop.List},
			}},
		}, &oop.Fn{
			Name:              "input",
			Src:               functions.Input,
			DefaultParamCount: 1,
			Params: []oop.Param{{
				Name:       "message",
				DefaultVal: oop.Val{Data: "", Type: oop.String},
			}},
		}, &oop.Fn{
			Name:              "exit",
			DefaultParamCount: 1,
			Src:               functions.Exit,
			Params: []oop.Param{{
				Name:       "code",
				DefaultVal: oop.Val{Data: "0", Type: oop.Int},
			}},
		}, &oop.Fn{
			Name:              "len",
			Src:               functions.Len,
			DefaultParamCount: 0,
			Params:            []oop.Param{{Name: "object"}},
		}, &oop.Fn{
			Name:              "range",
			DefaultParamCount: 1,
			Src:               functions.Range,
			Params: []oop.Param{
				{Name: "start"},
				{Name: "to"},
				{
					Name:       "step",
					DefaultVal: oop.Val{Data: "1", Type: oop.Int},
				},
			},
		}, &oop.Fn{
			Name:              "calloc",
			Src:               functions.Calloc,
			DefaultParamCount: 0,
			Params:            []oop.Param{{Name: "size"}},
		}, &oop.Fn{
			Name:              "realloc",
			DefaultParamCount: 0,
			Src:               functions.Realloc,
			Params:            []oop.Param{{Name: "base"}, {Name: "size"}},
		}, &oop.Fn{
			Name:              "string",
			Src:               functions.String,
			DefaultParamCount: 1,
			Params: []oop.Param{
				{Name: "object"},
				{
					Name:       "type",
					DefaultVal: oop.Val{Data: "parse", Type: oop.String},
				},
			},
		}, &oop.Fn{
			Name:              "int",
			Src:               functions.Int,
			DefaultParamCount: 1,
			Params: []oop.Param{
				{Name: "object"},
				{
					Name:       "type",
					DefaultVal: oop.Val{Data: "parse", Type: oop.String},
				},
			},
		}, &oop.Fn{
			Name:              "float",
			Src:               functions.Float,
			DefaultParamCount: 0,
			Params:            []oop.Param{{Name: "object"}},
		}, &oop.Fn{
			Name:              "panic",
			Src:               functions.Panic,
			DefaultParamCount: 0,
			Params:            []oop.Param{{Name: "msg"}},
		}, &oop.Fn{
			Name:              "type",
			Src:               functions.Type,
			DefaultParamCount: 0,
			Params:            []oop.Param{{Name: "obj"}},
		},
	)
}

func (p *Parser) processIf(tokens []obj.Token) uint8 {
	blockIndex := findBlock(tokens)
	blockTokens, conditionTokens := p.getBlock(tokens[blockIndex:]), tokens[1:blockIndex]
	// Condition is empty?
	if len(conditionTokens) == 0 {
		first := tokens[0]
		fract.IPanicC(first.File, first.Line, first.Column+len(first.Val), obj.SyntaxPanic, "Condition is empty!")
	}
	condition := p.prococessCondition(conditionTokens)
	varLen := len(p.defs.Vars)
	fnLen := len(p.defs.Funcs)
	impLen := len(p.packages)
	keywordState := fract.NA
	for _, tokens := range blockTokens {
		// Condition is true?
		if condition == "true" && keywordState == fract.NA {
			if keywordState = p.processExpression(tokens); keywordState != fract.NA {
				break
			}
		} else {
			break
		}
	}
rep:
	p.index++
	if p.index >= len(p.Tokens) {
		p.index--
		goto end
	}
	tokens = p.Tokens[p.index]
	if tokens[0].Type != fract.Else {
		p.index--
		goto end
	}
	if len(tokens) > 1 && tokens[1].Type == fract.If { // Else if.
		blockIndex = findBlock(tokens)
		blockTokens, conditionTokens = p.getBlock(tokens[blockIndex:]), tokens[2:blockIndex]
		// Condition is empty?
		if len(conditionTokens) == 0 {
			first := tokens[1]
			fract.IPanicC(first.File, first.Line, first.Column+len(first.Val), obj.SyntaxPanic, "Condition is empty!")
		}
		if condition == "true" {
			goto rep
		}
		condition = p.prococessCondition(conditionTokens)
		for _, tokens := range blockTokens {
			// Condition is true?
			if condition == "true" && keywordState == fract.NA {
				if keywordState = p.processExpression(tokens); keywordState != fract.NA {
					break
				}
			} else {
				break
			}
		}
		goto rep
	}
	blockTokens = p.getBlock(tokens[1:])
	if condition == "true" {
		goto end
	}
	for _, tokens := range blockTokens {
		// Condition is true?
		if keywordState == fract.NA {
			if keywordState = p.processExpression(tokens); keywordState != fract.NA {
				break
			}
		}
	}
end:
	p.defs.Vars = p.defs.Vars[:varLen]
	p.defs.Funcs = p.defs.Funcs[:fnLen]
	p.packages = p.packages[:impLen]
	return keywordState
}

// checkPublic name access.
func checkPublic(f *obj.File, name obj.Token) {
	if f != nil {
		if f == name.File {
			return
		}
	}
	if !unicode.IsUpper(rune(name.Val[0])) {
		fract.IPanic(name, obj.NamePanic, "Name is not defined: "+name.Val)
	}
}

func (p *Parser) processTryCatch(tokens []obj.Token) uint8 {
	fract.TryCount++
	var (
		varLen   = len(p.defs.Vars)
		fnLen    = len(p.defs.Funcs)
		impLen   = len(p.packages)
		deferLen = len(defers)
		kws      = fract.NA
	)
	b := &obj.Block{
		Try: func() {
			for _, tks := range p.getBlock(tokens[1:]) {
				if kws = p.processExpression(tks); kws != fract.NA {
					break
				}
			}
			if p.Tokens[p.index+1][0].Type == fract.Catch {
				p.index++
			}
			fract.TryCount--
			p.defs.Vars = p.defs.Vars[:varLen]
			p.defs.Funcs = p.defs.Funcs[:fnLen]
			p.packages = p.packages[:impLen]
			for index := len(defers) - 1; index >= deferLen; index-- {
				defers[index].Call()
			}
			defers = defers[:deferLen]
		},
		Catch: func(cp obj.Panic) {
			for index := len(defers) - 1; index >= deferLen; index-- {
				defers[index].Call()
			}
			p.loopCount = 0
			fract.TryCount--
			p.defs.Vars = p.defs.Vars[:varLen]
			p.defs.Funcs = p.defs.Funcs[:fnLen]
			p.packages = p.packages[:impLen]
			defers = defers[:deferLen]
			p.index++
			tokens = p.Tokens[p.index]
			if tokens[0].Type != fract.Catch {
				p.index--
				return
			}
			l := len(tokens)
			if l < 2 {
				fract.IPanic(tokens[0], obj.SyntaxPanic, "Invalid syntax!")
			}
			nameTk := tokens[1]
			if nameTk.Type == fract.Name {
				if ln := p.defIndexByName(nameTk.Val); ln != -1 {
					fract.IPanic(nameTk, obj.NamePanic, "\""+nameTk.Val+"\" is already defined at line: "+fmt.Sprint(ln))
				}
				p.defs.Vars = append(p.defs.Vars, oop.Var{
					Name: nameTk.Val,
					Line: nameTk.Line,
					Val:  oop.Val{Data: cp.Msg, Type: oop.String},
				})
			}
			var block [][]obj.Token
			if nameTk.Type == fract.Name {
				if l < 3 {
					fract.IPanic(tokens[1], obj.SyntaxPanic, "Invalid syntax!")
				}
				block = p.getBlock(tokens[2:])
			} else {
				block = p.getBlock(tokens[1:])
			}
			for _, tks := range block {
				if kws = p.processExpression(tks); kws != fract.NA {
					break
				}
			}
			p.defs.Vars = p.defs.Vars[:varLen]
			p.defs.Funcs = p.defs.Funcs[:fnLen]
			for i := len(defers) - 1; i >= deferLen; i-- {
				defers[i].Call()
			}
			defers = defers[:deferLen]
		},
	}
	b.Do()
	b = nil
	return kws
}

// TODO: Add match-case.
//! A change added here(especially added a code block) must also be compatible with "imports.go" and

// processExpression and returns keyword state.
func (p *Parser) processExpression(tks []obj.Token) uint8 {
	switch firstTk := tks[0]; firstTk.Type {
	case fract.Value, fract.Brace, fract.Name:
		if firstTk.Type == fract.Name {
			braceCount := 0
			for _, tk := range tks {
				if tk.Type == fract.Brace {
					switch tk.Val {
					case " {", "[", "(":
						braceCount++
					default:
						braceCount--
					}
				}
				if braceCount > 0 {
					continue
				}
				if tk.Type == fract.Operator {
					switch tk.Val {
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
		if val := p.processValTokens(tks); fract.InteractiveShell {
			if val.Print() {
				println()
			}
		}
	case fract.Var:
		p.vardec(tks)
	case fract.If:
		return p.processIf(tks)
	case fract.Loop:
		p.loopCount++
		state := p.processLoop(tks)
		p.loopCount--
		return state
	case fract.Break:
		if p.loopCount < 1 {
			fract.IPanic(firstTk, obj.SyntaxPanic, "Break keyword only used in loops!")
		}
		return fract.LOOPBreak
	case fract.Continue:
		if p.loopCount < 1 {
			fract.IPanic(firstTk, obj.SyntaxPanic, "Continue keyword only used in loops!")
		}
		return fract.LOOPContinue
	case fract.Ret:
		if p.funcCount < 1 {
			fract.IPanic(firstTk, obj.SyntaxPanic, "Return keyword only used in functions!")
		}
		if len(tks) > 1 {
			value := p.processValTokens(tks[1:])
			p.returnVal = value
		} else {
			p.returnVal = nil
		}
		return fract.FUNCReturn
	case fract.Fn:
		p.funcdec(tks)
	case fract.Try:
		return p.processTryCatch(tks)
	case fract.Import:
		p.processImport(tks)
	case fract.Macro:
		p.processPragma(tks)
	case fract.Struct:
		p.structdec(tks)
	case fract.Class:
		p.classdec(tks)
	case fract.Defer, fract.Go:
		if l := len(tks); l < 2 {
			fract.IPanic(tks[0], obj.SyntaxPanic, "Function is not given!")
		} else if t := tks[l-1]; t.Type != fract.Brace && t.Val != ")" {
			fract.IPanicC(tks[0].File, tks[0].Line, tks[0].Column+len(tks[0].Val), obj.SyntaxPanic, "Invalid syntax!")
		}
		var valTokens []obj.Token
		braceCount := 0
		for i := len(tks) - 1; i >= 0; i-- {
			tk := tks[i]
			if tk.Type != fract.Brace {
				continue
			}
			switch tk.Val {
			case ")":
				braceCount++
			case "(":
				braceCount--
			}
			if braceCount > 0 {
				continue
			}
			valTokens = tks[1:i]
			break
		}
		if len(valTokens) == 0 && braceCount == 0 {
			fract.IPanic(tks[1], obj.SyntaxPanic, "Invalid syntax!")
		}
		// Function call.
		val := p.processValuePart(valuePartInfo{tokens: valTokens})
		if val.Type != oop.Func {
			fract.IPanic(tks[len(valTokens)], obj.ValuePanic, "Value is not function!")
		}
		if firstTk.Type == fract.Defer {
			defers = append(defers, p.funcCallModel(val.Data.(*oop.Fn), tks[len(valTokens):]))
		} else {
			go p.funcCallModel(val.Data.(*oop.Fn), tks[len(valTokens):]).Call()
		}
	default:
		fract.IPanic(firstTk, obj.SyntaxPanic, "Invalid syntax!")
	}
	return fract.NA
}
