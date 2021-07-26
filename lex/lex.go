package lex

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"unicode"

	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
	"github.com/fract-lang/fract/pkg/str"
)

// Lexer of Fract.
type Lex struct {
	lastTk obj.Token

	F            *obj.File
	Col          int // Last column.
	Ln           int // Last line.
	Fin          bool
	RangeComment bool
	Braces       int
	Brackets     int
	Parentheses  int
}

// error thrown exception.
func (l Lex) error(msg string) {
	fmt.Printf("File: %s\nPosition: %d:%d\n", l.F.P, l.Ln, l.Col)
	if !l.RangeComment { // Ignore multiline comment error.
		fmt.Println("    " + strings.ReplaceAll(l.F.Lns[l.Ln-1], "\t", " "))
		fmt.Println(str.Full(4+l.Col-2, ' ') + "^")
	}
	fmt.Println(msg)
	panic(nil)
}

// Check expected bracket or like and returns true if require retokenize, returns false if not.
// Thrown exception is syntax error.
func (l *Lex) checkExpected(msg string) bool {
	if l.Fin {
		if l.F.P != "<stdin>" {
			l.Ln-- // Subtract for correct line number.
			l.error(msg)
		}
		return false
	}
	return true
}

// Next lex next line.
func (l *Lex) Next() obj.Tokens {
	var tks obj.Tokens
	if l.Fin {
		return tks
	}
tokenize:
	if l.lastTk.T != fract.StatementTerminator {
		// Restore to defaults.
		l.Col = 1
		l.lastTk.T = fract.None
		l.lastTk.Ln = 0
		l.lastTk.Col = 0
		l.lastTk.V = ""
	}
	// Tokenize line.
	tk := l.Token()
	for tk.T != fract.None {
		if tk.T == fract.StatementTerminator {
			if l.Parentheses == 0 && l.Braces == 0 && l.Brackets == 0 {
				break
			}
			l.Ln++
		}
		if !l.RangeComment && tk.T != fract.Ignore {
			tks = append(tks, tk)
			l.lastTk = tk
		}
		tk = l.Token()
	}
	l.lastTk = tk
	l.Ln++
	l.Fin = l.Ln > len(l.F.Lns)
	switch {
	case l.Parentheses > 0:
		if l.checkExpected("Parentheses is expected to close...") {
			goto tokenize
		}
	case l.Braces > 0:
		if l.checkExpected("Brace is expected to close...") {
			goto tokenize
		}
	case l.Brackets > 0:
		if l.checkExpected("Bracket is expected to close...") {
			goto tokenize
		}
	case l.RangeComment:
		if l.checkExpected("Multiline comment is expected to close...") {
			goto tokenize
		}
	}
	return tks
}

var (
	numRgx  = *regexp.MustCompile(`^(-|)((\d+((\.\d+)|(\.\d+)?(e|E)(\-|\+)\d+)?)|(0x[[:xdigit:]]+))(\s|[[:punct:]]|$)`)
	nameRgx = *regexp.MustCompile(`^(-|)([\p{L}|_])([\p{L}0-9_]+)?([[:punct:]]|\s|$)`)
)

// isKeyword returns true if part is keyword, false if not.
func isKeyword(ln, kw string) bool {
	return regexp.MustCompile("^" + kw + `(\s+|$|[[:punct:]])`).MatchString(ln)
}

// getName returns name if next token is name, returns empty string if not.
func getName(ln string) string { return nameRgx.FindString(ln) }

// getNumeric returns numeric if next token is numeric, returns empty string if not.
func getNumeric(ln string) string { return numRgx.FindString(ln) }

// Process string espace sequence.
func (l *Lex) strseq(sb *strings.Builder, fln string) bool {
	// Is not espace sequence?
	if fln[l.Col-1] != '\\' {
		return false
	}
	l.Col++
	if l.Col >= len(fln)+1 {
		l.error("Charray literal is not defined full!")
	}
	switch fln[l.Col-1] {
	case '\\':
		sb.WriteByte('\\')
	case '"':
		sb.WriteByte('"')
	case '\'':
		sb.WriteByte('\'')
	case 'n':
		sb.WriteByte('\n')
	case 'r':
		sb.WriteByte('\r')
	case 't':
		sb.WriteByte('\t')
	case 'b':
		sb.WriteByte('\b')
	case 'f':
		sb.WriteByte('\f')
	case 'a':
		sb.WriteByte('\a')
	case 'v':
		sb.WriteByte('\v')
	default:
		l.error("Invalid escape sequence!")
	}
	return true
}

func (l *Lex) lexstr(tk *obj.Token, quote byte, fln string) {
	sb := new(strings.Builder)
	sb.WriteByte(quote)
	l.Col++
	for ; l.Col < len(fln)+1; l.Col++ {
		c := fln[l.Col-1]
		if c == quote { // Finish?
			sb.WriteByte(c)
			break
		} else if !l.strseq(sb, fln) {
			sb.WriteByte(c)
		}
	}
	tk.V = sb.String()
	if tk.V[len(tk.V)-1] != quote {
		l.error("Close quote is not found!")
	}
	tk.T = fract.Value
	l.Col -= sb.Len() - 1
}

func (l *Lex) lexname(tk *obj.Token, chk string) bool {
	// Remove punct.
	if chk[len(chk)-1] != '_' {
		r, _ := regexp.MatchString(`(\s|[[:punct:]])$`, chk)
		if r {
			chk = chk[:len(chk)-1]
		}
	}
	tk.V = chk
	tk.T = fract.Name
	return true
}

// Generate next token.
func (l *Lex) Token() obj.Token {
	tk := obj.Token{T: fract.None, F: l.F}

	fln := l.F.Lns[l.Ln-1] // Full line.
	// Line is finished.
	if l.Col > len(fln) {
		if l.RangeComment {
			l.F.Lns[l.Ln-1] = ""
		}
		return tk
	}
	// Resume.
	ln := fln[l.Col-1:]
	// Skip spaces.
	for i, c := range ln {
		if unicode.IsSpace(c) {
			l.Col++
			continue
		}
		ln = ln[i:]
		break
	}
	// Content is empty.
	if ln == "" {
		return tk
	}
	// Set token values.
	tk.Col = l.Col
	tk.Ln = l.Ln

	// ************
	//   Tokenize
	// ************

	if l.RangeComment { // Range comment.
		tk.T = fract.Ignore
		if strings.HasPrefix(ln, "*/") { // Range comment close.
			l.RangeComment = false
			l.Col += 2 // len("<#")
			return tk
		}
		l.Col++
		return tk
	}

	switch chk := getNumeric(ln); {
	case (chk != "" &&
		(l.lastTk.V == "" || l.lastTk.T == fract.Operator ||
			(l.lastTk.T == fract.Brace && l.lastTk.V != "]") ||
			l.lastTk.T == fract.StatementTerminator || l.lastTk.T == fract.Loop ||
			l.lastTk.T == fract.Comma || l.lastTk.T == fract.In || l.lastTk.T == fract.If ||
			l.lastTk.T == fract.Else || l.lastTk.T == fract.Ret || l.lastTk.T == fract.Colon)) ||
		isKeyword(ln, "NaN"): // Numeric oop.
		if chk == "" {
			chk = "NaN"
			l.Col += 3
		} else {
			// Remove punct.
			if lst := chk[len(chk)-1]; lst < '0' || lst > '9' {
				chk = chk[:len(chk)-1]
			}
			l.Col += len(chk)
			if strings.HasPrefix(chk, "0x") {
				// Parse hexadecimal to decimal.
				bi := new(big.Int)
				bi.SetString(chk[2:], 16)
				chk = bi.String()
			} else {
				// Parse floating-point.
				bf := new(big.Float)
				_, f := bf.SetString(chk)
				if !f {
					chk = bf.String()
				}
			}
		}
		tk.V = chk
		tk.T = fract.Value
		return tk
	case strings.HasPrefix(ln, "//"):
		l.F.Lns[l.Ln-1] = l.F.Lns[l.Ln-1][:l.Col-1] // Remove comment from original line.
		return tk
	case strings.HasPrefix(ln, "/*"):
		l.RangeComment = true
		tk.V = "/*"
		tk.T = fract.Ignore
	case ln[0] == '#':
		tk.V = "#"
		tk.T = fract.Macro
	case ln[0] == '\'':
		l.lexstr(&tk, '\'', fln)
	case ln[0] == '"':
		l.lexstr(&tk, '"', fln)
	case ln[0] == '.':
		tk.V = "."
		tk.T = fract.Dot
	case ln[0] == ';':
		tk.V = ";"
		tk.T = fract.StatementTerminator
		l.Ln--
	case strings.HasPrefix(ln, ":="):
		tk.V = ":="
		tk.T = fract.Operator
	case strings.HasPrefix(ln, "+="):
		tk.V = "+="
		tk.T = fract.Operator
	case strings.HasPrefix(ln, "**="):
		tk.V = "**="
		tk.T = fract.Operator
	case strings.HasPrefix(ln, "*="):
		tk.V = "*="
		tk.T = fract.Operator
	case strings.HasPrefix(ln, "/="):
		tk.V = "/="
		tk.T = fract.Operator
	case strings.HasPrefix(ln, "%="):
		tk.V = "%="
		tk.T = fract.Operator
	case strings.HasPrefix(ln, "-="):
		tk.V = "-="
		tk.T = fract.Operator
	case strings.HasPrefix(ln, "<<="):
		tk.V = "<<="
		tk.T = fract.Operator
	case strings.HasPrefix(ln, ">>="):
		tk.V = ">>="
		tk.T = fract.Operator
	case strings.HasPrefix(ln, "|="):
		tk.V = "|="
		tk.T = fract.Operator
	case strings.HasPrefix(ln, "^="):
		tk.V = "^="
		tk.T = fract.Operator
	case strings.HasPrefix(ln, "&="):
		tk.V = "&="
		tk.T = fract.Operator
	case ln[0] == '+':
		tk.V = "+"
		tk.T = fract.Operator
	case ln[0] == '-':
		// Check variable name.
		if check := getName(ln); check != "" { // Name.
			if !l.lexname(&tk, check) {
				return tk
			}
			break
		}
		tk.V = "-"
		tk.T = fract.Operator
	case ln[0] == ':':
		tk.V = ":"
		tk.T = fract.Colon
	case strings.HasPrefix(ln, "**"):
		tk.V = "**"
		tk.T = fract.Operator
	case ln[0] == '*':
		tk.V = "*"
		tk.T = fract.Operator
	case ln[0] == '/':
		tk.V = "/"
		tk.T = fract.Operator
	case ln[0] == '%':
		tk.V = "%"
		tk.T = fract.Operator
	case ln[0] == '(':
		l.Parentheses++
		tk.V = "("
		tk.T = fract.Brace
	case ln[0] == ')':
		l.Parentheses--
		if l.Parentheses < 0 {
			l.error("The extra parentheses are closed!")
		}
		tk.V = ")"
		tk.T = fract.Brace
	case ln[0] == '{':
		l.Braces++
		tk.V = "{"
		tk.T = fract.Brace
	case ln[0] == '}':
		l.Braces--
		if l.Braces < 0 {
			l.error("The extra brace are closed!")
		}
		tk.V = "}"
		tk.T = fract.Brace
	case ln[0] == '[':
		l.Brackets++
		tk.V = "["
		tk.T = fract.Brace
	case ln[0] == ']':
		l.Brackets--
		if l.Brackets < 0 {
			l.error("The extra bracket are closed!")
		}
		tk.V = "]"
		tk.T = fract.Brace
	case strings.HasPrefix(ln, "<<"):
		tk.V = "<<"
		tk.T = fract.Operator
	case strings.HasPrefix(ln, ">>"):
		tk.V = ">>"
		tk.T = fract.Operator
	case ln[0] == ',':
		tk.V = ","
		tk.T = fract.Comma
	case strings.HasPrefix(ln, "&&"):
		tk.V = "&&"
		tk.T = fract.Operator
	case strings.HasPrefix(ln, "||"):
		tk.V = "||"
		tk.T = fract.Operator
	case ln[0] == '|':
		tk.V = "|"
		tk.T = fract.Operator
	case ln[0] == '&':
		tk.V = "&"
		tk.T = fract.Operator
	case ln[0] == '^':
		tk.V = "^"
		tk.T = fract.Operator
	case strings.HasPrefix(ln, ">="):
		tk.V = ">="
		tk.T = fract.Operator
	case strings.HasPrefix(ln, "<="):
		tk.V = "<="
		tk.T = fract.Operator
	case strings.HasPrefix(ln, "=="):
		tk.V = "=="
		tk.T = fract.Operator
	case strings.HasPrefix(ln, "<>"):
		tk.V = "<>"
		tk.T = fract.Operator
	case ln[0] == '>':
		tk.V = ">"
		tk.T = fract.Operator
	case ln[0] == '<':
		tk.V = "<"
		tk.T = fract.Operator
	case ln[0] == '=':
		tk.V = "="
		tk.T = fract.Operator
	case strings.HasPrefix(ln, "..."):
		tk.V = "..."
		tk.T = fract.Params
	case isKeyword(ln, "var"):
		tk.V = "var"
		tk.T = fract.Var
	case isKeyword(ln, "mut"):
		tk.V = "mut"
		tk.T = fract.Var
	case isKeyword(ln, "const"):
		tk.V = "const"
		tk.T = fract.Var
	case isKeyword(ln, "defer"):
		tk.V = "defer"
		tk.T = fract.Defer
	case isKeyword(ln, "if"):
		tk.V = "if"
		tk.T = fract.If
	case isKeyword(ln, "else"):
		tk.V = "else"
		tk.T = fract.Else
	case isKeyword(ln, "for"):
		tk.V = "for"
		tk.T = fract.Loop
	case isKeyword(ln, "in"):
		tk.V = "in"
		tk.T = fract.In
	case isKeyword(ln, "break"):
		tk.V = "break"
		tk.T = fract.Break
	case isKeyword(ln, "continue"):
		tk.V = "continue"
		tk.T = fract.Continue
	case isKeyword(ln, "func"):
		tk.V = "func"
		tk.T = fract.Func
	case isKeyword(ln, "ret"):
		tk.V = "ret"
		tk.T = fract.Ret
	case isKeyword(ln, "try"):
		tk.V = "try"
		tk.T = fract.Try
	case isKeyword(ln, "catch"):
		tk.V = "catch"
		tk.T = fract.Catch
	case isKeyword(ln, "open"):
		tk.V = "open"
		tk.T = fract.Import
	case isKeyword(ln, "true"):
		tk.V = "true"
		tk.T = fract.Value
	case isKeyword(ln, "false"):
		tk.V = "false"
		tk.T = fract.Value
	case isKeyword(ln, "go"):
		tk.V = "go"
		tk.T = fract.Go
	case isKeyword(ln, "package"):
		tk.V = "package"
		tk.T = fract.Package
	case isKeyword(ln, "struct"):
		tk.V = "struct"
		tk.T = fract.Struct
	case isKeyword(ln, "class"):
		tk.V = "class"
		tk.T = fract.Class
	default: // Alternates
		// Check variable name.
		if chk := getName(ln); chk != "" { // Name.
			if !l.lexname(&tk, chk) {
				return tk
			}
		} else {
			l.error("Invalid token!")
		}
	}
	l.Col += len(tk.V)
	return tk
}
