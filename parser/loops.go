package parser

import (
	"fmt"

	"github.com/fract-lang/fract/oop"
	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
)

// Loop.
type loop struct {
	a         oop.Val
	b         oop.Val
	val       oop.Val
	breakLoop bool
}

func (l *loop) run(b func()) {
	switch l.val.Type {
	case oop.List:
		l.a.Type = oop.Int
		for i, e := range l.val.Data.(*oop.ListModel).Elems {
			l.a.Data = float64(i)
			l.b = e
			b()
			if l.breakLoop {
				break
			}
		}
	case oop.String:
		l.a.Type = oop.Int
		l.b.Type = oop.String
		for i, e := range l.val.Data.(string) {
			l.a.Data = float64(i)
			l.b.Data = string(e)
			b()
			if l.breakLoop {
				break
			}
		}
	case oop.Map:
		for k, v := range l.val.Data.(oop.MapModel).Map {
			l.a = k
			l.b = v
			b()
			if l.breakLoop {
				break
			}
		}
	}
}

// Returns kwstate's return format.
func processKeywordState(kws uint8) uint8 {
	if kws != fract.FUNCReturn {
		return fract.NA
	}
	return kws
}

func (p *Parser) processLoop(tokens []obj.Token) uint8 {
	blockIndex := findBlock(tokens)
	blockTokens, tokens := p.getBlock(tokens[blockIndex:]), tokens[1:blockIndex]
	funcLen := len(p.defs.Funcs)
	impLen := len(p.packages)
	breakLoop := false
	keywordState := fract.NA
	parserTokens := p.Tokens
	parserIndex := p.index
	//*************
	//    WHILE
	//*************
	if len(tokens) == 0 || len(tokens) >= 1 {
		if len(tokens) == 0 || len(tokens) == 1 || len(tokens) >= 1 && tokens[1].Type != fract.In && tokens[1].Type != fract.Comma {
			varLen := len(p.defs.Vars)
			// Infinity loop.
			if len(tokens) == 0 {
			infinity:
				p.Tokens = blockTokens
				for p.index = 0; p.index < len(p.Tokens); p.index++ {
					keywordState = p.processExpression(p.Tokens[p.index])
					if keywordState == fract.LOOPBreak || keywordState == fract.FUNCReturn { // Break loop or return.
						p.Tokens = parserTokens
						p.index = parserIndex
						return processKeywordState(keywordState)
					} else if keywordState == fract.LOOPContinue { // Continue loop.
						break
					}
				}
				// Remove temporary variables.
				p.defs.Vars = p.defs.Vars[:varLen]
				// Remove temporary functions.
				p.defs.Funcs = p.defs.Funcs[:funcLen]
				// Remove temporary imports.
				p.packages = p.packages[:impLen]
				goto infinity
			}
		while:
			// Interpret/skip block.
			condition := p.prococessCondition(tokens)
			p.Tokens = blockTokens
			for p.index = 0; p.index < len(p.Tokens); p.index++ {
				// Condition is true?
				if condition {
					keywordState = p.processExpression(p.Tokens[p.index])
					if keywordState == fract.LOOPBreak || keywordState == fract.FUNCReturn { // Break loop or return.
						breakLoop = true
						break
					} else if keywordState == fract.LOOPContinue { // Continue loop.
						break
					}
				} else {
					breakLoop = true
					break
				}
			}
			// Remove temporary variables.
			p.defs.Vars = p.defs.Vars[:varLen]
			// Remove temporary functions.
			p.defs.Funcs = p.defs.Funcs[:funcLen]
			// Remove temporary imports.
			p.packages = p.packages[:impLen]
			condition = p.prococessCondition(tokens)
			if breakLoop || !condition {
				p.Tokens = parserTokens
				p.index = parserIndex
				return processKeywordState(keywordState)
			}
			goto while
		}
	}
	//*************
	//   FOREACH
	//*************
	nameTK := tokens[0]
	// Name is not name?
	if nameTK.Type != fract.Name {
		fract.IPanic(nameTK, obj.SyntaxPanic, "This is not a valid name!")
	}
	if nameTK.Val != "_" {
		if !isValidName(nameTK.Val) {
			fract.IPanic(nameTK, obj.NamePanic, "Invalid name!")
		}
		if ln := p.defIndexByName(nameTK.Val); ln != -1 {
			fract.IPanic(nameTK, obj.NamePanic, "\""+nameTK.Val+"\" is already defined at line: "+fmt.Sprint(ln))
		}
	} else {
		nameTK.Val = ""
	}
	// Element name?
	elemName := ""
	if tokens[1].Type == fract.Comma {
		if len(tokens) < 3 || tokens[2].Type != fract.Name {
			fract.IPanic(tokens[1], obj.SyntaxPanic, "Element name is not defined!")
		}
		if tokens[2].Val != "_" {
			elemName = tokens[2].Val
			if !isValidName(elemName) {
				fract.IPanic(tokens[2], obj.NamePanic, "Invalid name!")
			}
			if ln := p.defIndexByName(tokens[2].Val); ln != -1 {
				fract.IPanic(tokens[2], obj.NamePanic, "\""+elemName+"\" is already defined at line: "+fmt.Sprint(ln))
			}
		}
		if len(tokens)-3 == 0 {
			tokens[2].Column += len(tokens[2].Val)
			fract.IPanic(tokens[2], obj.SyntaxPanic, "Value is not given!")
		}
		tokens = tokens[2:]
	}
	if len(tokens) < 3 {
		fract.IPanic(tokens[1], obj.SyntaxPanic, "Value is not given!")
	} else if t := tokens[1]; t.Type != fract.In {
		fract.IPanic(tokens[1], obj.SyntaxPanic, "Invalid syntax!")
	}
	tokens = tokens[2:]
	val := *p.processValTokens(tokens)
	// Type is not list?
	if !val.IsEnum() {
		fract.IPanic(tokens[0], obj.ValuePanic, "Foreach loop must defined enumerable value!")
	}
	index := &oop.Var{Name: nameTK.Val, Val: oop.Val{Data: "0", Type: oop.Int}}
	element := &oop.Var{Name: elemName}
	p.defs.Vars = append(p.defs.Vars, index, element)
	vars := p.defs.Vars
	// Interpret block.
	l := loop{val: val}
	l.run(func() {
		index.Val = l.a
		element.Val = l.b
		p.Tokens = blockTokens
		for p.index = 0; p.index < len(p.Tokens); p.index++ {
			keywordState = p.processExpression(p.Tokens[p.index])
			if keywordState == fract.LOOPBreak || keywordState == fract.FUNCReturn { // Break loop or return.
				breakLoop = true
				break
			} else if keywordState == fract.LOOPContinue { // Continue loop.
				break
			}
		}
		// Remove temporary variables.
		p.defs.Vars = vars
		// Remove temporary functions.
		p.defs.Funcs = p.defs.Funcs[:funcLen]
		// Remove temporary imports.
		p.packages = p.packages[:impLen]
		l.breakLoop = breakLoop
	})
	p.Tokens = parserTokens
	p.index = parserIndex
	// Remove loop variables.
	index = nil
	element = nil
	p.defs.Vars = vars[:len(vars)-2]
	return processKeywordState(keywordState)
}
