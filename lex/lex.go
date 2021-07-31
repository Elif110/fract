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

	File         *obj.File
	Column       int // Last column.
	Line         int // Last line.
	Finished     bool
	RangeComment bool
	Braces       int
	Brackets     int
	Parentheses  int
}

// error thrown exception.
func (l Lex) error(msg string) {
	fmt.Printf("File: %s\nPosition: %d:%d\n", l.File.Path, l.Line, l.Column)
	if !l.RangeComment { // Ignore multiline comment error.
		fmt.Println("    " + strings.ReplaceAll(l.File.Lines[l.Line-1], "\t", " "))
		fmt.Println(str.Full(4+l.Column-2, ' ') + "^")
	}
	fmt.Println(msg)
	panic(nil)
}

// Check expected bracket or like and returns true if require retokenize, returns false if not.
// Thrown exception is syntax error.
func (l *Lex) checkExpected(msg string) bool {
	if l.Finished {
		if l.File.Path != "<stdin>" {
			l.Line-- // Subtract for correct line number.
			l.error(msg)
		}
		return false
	}
	return true
}

// Next lex next line.
func (l *Lex) Next() []obj.Token {
	var tokens []obj.Token
	if l.Finished {
		return tokens
	}
tokenize:
	if l.lastTk.Type != fract.StatementTerminator {
		// Restore to defaults.
		l.Column = 1
		l.lastTk.Type = fract.NA
		l.lastTk.Line = 0
		l.lastTk.Column = 0
		l.lastTk.Val = ""
	}
	// Tokenize line.
	tk := l.Token()
	for tk.Type != fract.NA {
		if tk.Type == fract.StatementTerminator {
			if l.Parentheses == 0 && l.Braces == 0 && l.Brackets == 0 {
				break
			}
			l.Line++
		}
		if !l.RangeComment && tk.Type != fract.Ignore {
			tokens = append(tokens, tk)
			l.lastTk = tk
		}
		tk = l.Token()
	}
	l.lastTk = tk
	l.Line++
	l.Finished = l.Line > len(l.File.Lines)
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
	return tokens
}

var (
	numRgx  = *regexp.MustCompile(`^(-|)((\d+((\.\d+)|(\.\d+)?(e|E)(\-|\+)\d+)?)|(0x[[:xdigit:]]+))(\s|[[:punct:]]|$)`)
	nameRgx = *regexp.MustCompile(`^[\p{L}|_]([\p{L}0-9_]+)?([[:punct:]]|\s|$)`)
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
func (l *Lex) strseq(sb *strings.Builder, fullLn string) bool {
	// Is not espace sequence?
	if fullLn[l.Column-1] != '\\' {
		return false
	}
	l.Column++
	if l.Column >= len(fullLn)+1 {
		l.error("String literal is not defined full!")
	}
	switch fullLn[l.Column-1] {
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

func (l *Lex) lexString(tk *obj.Token, quote byte, fullLn string) {
	sb := new(strings.Builder)
	sb.WriteByte(quote)
	l.Column++
	for ; l.Column < len(fullLn)+1; l.Column++ {
		c := fullLn[l.Column-1]
		if c == quote { // Finish?
			sb.WriteByte(c)
			break
		} else if !l.strseq(sb, fullLn) {
			sb.WriteByte(c)
		}
	}
	tk.Val = sb.String()
	if tk.Val[len(tk.Val)-1] != quote {
		l.error("Close quote is not found!")
	}
	tk.Type = fract.Value
	l.Column -= sb.Len() - 1
}

func (l *Lex) lexname(tk *obj.Token, chk string) bool {
	// Remove punct.
	if chk[len(chk)-1] != '_' {
		r, _ := regexp.MatchString(`(\s|[[:punct:]])$`, chk)
		if r {
			chk = chk[:len(chk)-1]
		}
	}
	tk.Val = chk
	tk.Type = fract.Name
	return true
}

// Generate next token.
func (l *Lex) Token() obj.Token {
	tk := obj.Token{Type: fract.NA, File: l.File}

	fullLn := l.File.Lines[l.Line-1] // Full line.
	// Line is finished.
	if l.Column > len(fullLn) {
		if l.RangeComment {
			l.File.Lines[l.Line-1] = ""
		}
		return tk
	}
	// Resume.
	ln := fullLn[l.Column-1:]
	// Skip spaces.
	for i, r := range ln {
		if unicode.IsSpace(r) {
			l.Column++
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
	tk.Column = l.Column
	tk.Line = l.Line

	// ************
	//   Tokenize
	// ************
	if l.RangeComment { // Range comment.
		tk.Type = fract.Ignore
		if strings.HasPrefix(ln, "*/") { // Range comment close.
			l.RangeComment = false
			l.Column += 2 // len("<#")
			return tk
		}
		l.Column++
		return tk
	}
	switch chk := getNumeric(ln); {
	case (chk != "" && (l.lastTk.Val == "" || l.lastTk.Type == fract.Operator ||
		(l.lastTk.Type == fract.Brace && l.lastTk.Val != "}" && l.lastTk.Val != "]" && l.lastTk.Val != ")") ||
		l.lastTk.Type == fract.StatementTerminator || l.lastTk.Type == fract.Loop ||
		l.lastTk.Type == fract.Comma || l.lastTk.Type == fract.In || l.lastTk.Type == fract.If ||
		l.lastTk.Type == fract.Else || l.lastTk.Type == fract.Ret || l.lastTk.Type == fract.Colon)) ||
		isKeyword(ln, "NaN"): // Numeric oop.
		if chk == "" {
			chk = "NaN"
			l.Column += 3
		} else {
			// Remove punct.
			if lst := chk[len(chk)-1]; lst < '0' || lst > '9' {
				chk = chk[:len(chk)-1]
			}
			l.Column += len(chk)
			if strings.HasPrefix(chk, "0x") {
				// Parse hexadecimal to decimal.
				bigInt := new(big.Int)
				bigInt.SetString(chk[2:], 16)
				chk = bigInt.String()
			} else {
				// Parse floating-point.
				bigFloat := new(big.Float)
				_, f := bigFloat.SetString(chk)
				if !f {
					chk = bigFloat.String()
				}
			}
		}
		tk.Val = chk
		tk.Type = fract.Value
		return tk
	case strings.HasPrefix(ln, "//"):
		l.File.Lines[l.Line-1] = l.File.Lines[l.Line-1][:l.Column-1] // Remove comment from original line.
		return tk
	case strings.HasPrefix(ln, "/*"):
		l.RangeComment = true
		tk.Val = "/*"
		tk.Type = fract.Ignore
	case ln[0] == '#':
		tk.Val = "#"
		tk.Type = fract.Macro
	case ln[0] == '\'':
		l.lexString(&tk, '\'', fullLn)
	case ln[0] == '"':
		l.lexString(&tk, '"', fullLn)
	case ln[0] == ';':
		tk.Val = ";"
		tk.Type = fract.StatementTerminator
		l.Line--
	case strings.HasPrefix(ln, ":="):
		tk.Val = ":="
		tk.Type = fract.Operator
	case strings.HasPrefix(ln, "+="):
		tk.Val = "+="
		tk.Type = fract.Operator
	case strings.HasPrefix(ln, "**="):
		tk.Val = "**="
		tk.Type = fract.Operator
	case strings.HasPrefix(ln, "*="):
		tk.Val = "*="
		tk.Type = fract.Operator
	case strings.HasPrefix(ln, "/="):
		tk.Val = "/="
		tk.Type = fract.Operator
	case strings.HasPrefix(ln, "%="):
		tk.Val = "%="
		tk.Type = fract.Operator
	case strings.HasPrefix(ln, "-="):
		tk.Val = "-="
		tk.Type = fract.Operator
	case strings.HasPrefix(ln, "<<="):
		tk.Val = "<<="
		tk.Type = fract.Operator
	case strings.HasPrefix(ln, ">>="):
		tk.Val = ">>="
		tk.Type = fract.Operator
	case strings.HasPrefix(ln, "|="):
		tk.Val = "|="
		tk.Type = fract.Operator
	case strings.HasPrefix(ln, "^="):
		tk.Val = "^="
		tk.Type = fract.Operator
	case strings.HasPrefix(ln, "&="):
		tk.Val = "&="
		tk.Type = fract.Operator
	case ln[0] == '+':
		tk.Val = "+"
		tk.Type = fract.Operator
	case ln[0] == '-':
		tk.Val = "-"
		tk.Type = fract.Operator
	case ln[0] == ':':
		tk.Val = ":"
		tk.Type = fract.Colon
	case strings.HasPrefix(ln, "**"):
		tk.Val = "**"
		tk.Type = fract.Operator
	case ln[0] == '*':
		tk.Val = "*"
		tk.Type = fract.Operator
	case ln[0] == '/':
		tk.Val = "/"
		tk.Type = fract.Operator
	case ln[0] == '%':
		tk.Val = "%"
		tk.Type = fract.Operator
	case ln[0] == '(':
		l.Parentheses++
		tk.Val = "("
		tk.Type = fract.Brace
	case ln[0] == ')':
		l.Parentheses--
		if l.Parentheses < 0 {
			l.error("The extra parentheses are closed!")
		}
		tk.Val = ")"
		tk.Type = fract.Brace
	case ln[0] == '{':
		l.Braces++
		tk.Val = "{"
		tk.Type = fract.Brace
	case ln[0] == '}':
		l.Braces--
		if l.Braces < 0 {
			l.error("The extra brace are closed!")
		}
		tk.Val = "}"
		tk.Type = fract.Brace
	case ln[0] == '[':
		l.Brackets++
		tk.Val = "["
		tk.Type = fract.Brace
	case ln[0] == ']':
		l.Brackets--
		if l.Brackets < 0 {
			l.error("The extra bracket are closed!")
		}
		tk.Val = "]"
		tk.Type = fract.Brace
	case strings.HasPrefix(ln, "<<"):
		tk.Val = "<<"
		tk.Type = fract.Operator
	case strings.HasPrefix(ln, ">>"):
		tk.Val = ">>"
		tk.Type = fract.Operator
	case ln[0] == ',':
		tk.Val = ","
		tk.Type = fract.Comma
	case strings.HasPrefix(ln, "&&"):
		tk.Val = "&&"
		tk.Type = fract.Operator
	case strings.HasPrefix(ln, "||"):
		tk.Val = "||"
		tk.Type = fract.Operator
	case ln[0] == '|':
		tk.Val = "|"
		tk.Type = fract.Operator
	case ln[0] == '&':
		tk.Val = "&"
		tk.Type = fract.Operator
	case ln[0] == '^':
		tk.Val = "^"
		tk.Type = fract.Operator
	case strings.HasPrefix(ln, ">="):
		tk.Val = ">="
		tk.Type = fract.Operator
	case strings.HasPrefix(ln, "<="):
		tk.Val = "<="
		tk.Type = fract.Operator
	case strings.HasPrefix(ln, "=="):
		tk.Val = "=="
		tk.Type = fract.Operator
	case strings.HasPrefix(ln, "!="):
		tk.Val = "!="
		tk.Type = fract.Operator
	case ln[0] == '>':
		tk.Val = ">"
		tk.Type = fract.Operator
	case ln[0] == '<':
		tk.Val = "<"
		tk.Type = fract.Operator
	case ln[0] == '=':
		tk.Val = "="
		tk.Type = fract.Operator
	case strings.HasPrefix(ln, "..."):
		tk.Val = "..."
		tk.Type = fract.Params
	case ln[0] == '.':
		tk.Val = "."
		tk.Type = fract.Dot
	case isKeyword(ln, "var"):
		tk.Val = "var"
		tk.Type = fract.Var
	case isKeyword(ln, "mut"):
		tk.Val = "mut"
		tk.Type = fract.Var
	case isKeyword(ln, "const"):
		tk.Val = "const"
		tk.Type = fract.Var
	case isKeyword(ln, "defer"):
		tk.Val = "defer"
		tk.Type = fract.Defer
	case isKeyword(ln, "if"):
		tk.Val = "if"
		tk.Type = fract.If
	case isKeyword(ln, "else"):
		tk.Val = "else"
		tk.Type = fract.Else
	case isKeyword(ln, "for"):
		tk.Val = "for"
		tk.Type = fract.Loop
	case isKeyword(ln, "in"):
		tk.Val = "in"
		tk.Type = fract.In
	case isKeyword(ln, "break"):
		tk.Val = "break"
		tk.Type = fract.Break
	case isKeyword(ln, "continue"):
		tk.Val = "continue"
		tk.Type = fract.Continue
	case isKeyword(ln, "fn"):
		tk.Val = "fn"
		tk.Type = fract.Fn
	case isKeyword(ln, "ret"):
		tk.Val = "ret"
		tk.Type = fract.Ret
	case isKeyword(ln, "try"):
		tk.Val = "try"
		tk.Type = fract.Try
	case isKeyword(ln, "catch"):
		tk.Val = "catch"
		tk.Type = fract.Catch
	case isKeyword(ln, "open"):
		tk.Val = "open"
		tk.Type = fract.Import
	case isKeyword(ln, "true"):
		tk.Val = "true"
		tk.Type = fract.Value
	case isKeyword(ln, "false"):
		tk.Val = "false"
		tk.Type = fract.Value
	case isKeyword(ln, "go"):
		tk.Val = "go"
		tk.Type = fract.Go
	case isKeyword(ln, "package"):
		tk.Val = "package"
		tk.Type = fract.Package
	case isKeyword(ln, "struct"):
		tk.Val = "struct"
		tk.Type = fract.Struct
	case isKeyword(ln, "class"):
		tk.Val = "class"
		tk.Type = fract.Class
	case isKeyword(ln, "none"):
		tk.Val = "none"
		tk.Type = fract.None
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
	l.Column += len(tk.Val)
	return tk
}
