package parser

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/fract-lang/fract/oop"
	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
	"github.com/fract-lang/fract/pkg/str"
)

func compareValues(operator string, left, right oop.Val) bool {
	if left.Type != right.Type && (left.Type == oop.String || right.Type == oop.String) {
		return false
	}
	switch operator {
	case "==":
		if !left.Equals(right) {
			return false
		}
	case "!=":
		if !left.NotEquals(right) {
			return false
		}
	case ">":
		if !left.Greater(right) {
			return false
		}
	case "<":
		if !left.Less(right) {
			return false
		}
	case ">=":
		if !left.GreaterEquals(right) {
			return false
		}
	case "<=":
		if !left.LessEquals(right) {
			return false
		}
	}
	return true
}

func compare(left, right oop.Val, operator obj.Token) bool {
	if operator.Val == "in" {
		if !right.IsEnum() {
			fract.IPanic(operator, obj.ValuePanic, "Value is should be enumerable!")
		}
		switch right.Type {
		case oop.List:
			leftStr := left.String()
			for _, elem := range right.Data.(*oop.ListModel).Elems {
				if strings.Contains(elem.String(), leftStr) {
					return true
				}
			}
			return false
		case oop.Map:
			_, ok := right.Data.(oop.MapModel).Map[left]
			return ok
		}
		// String.
		if left.Type == oop.List {
			rightStr := right.String()
			for _, elem := range left.Data.(*oop.ListModel).Elems {
				if elem.Type != oop.String {
					fract.IPanic(operator, obj.ValuePanic, "All values is not string!")
				}
				if strings.Contains(rightStr, elem.String()) {
					return true
				}
			}
		} else {
			if right.Type != oop.String {
				fract.IPanic(operator, obj.ValuePanic, "All datas is not string!")
			}
			if strings.Contains(right.String(), left.String()) {
				return true
			}
		}
		return false
	}
	return compareValues(operator.Val, left, right)
}

func (p *Parser) prococessCondition(tokens []obj.Token) string {
	trueVal := oop.Val{Data: "true", Type: oop.Bool}
	// Process condition.
	orParts := conditionalProcesses(tokens, "||")
	for _, or := range orParts {
		// Decompose and conditions.
		andParts := conditionalProcesses(or, "&&")
		// Is and long statement?
		if len(andParts) > 1 {
			for _, and := range andParts {
				index, operator := findConditionOperator(and)
				// Operator is not found?
				if index == -1 {
					operator.Val = "=="
					if compare(*p.processValTokens(and), trueVal, operator) {
						return "true"
					}
					return "false"
				}
				// Operator is first or last?
				if index == 0 {
					fract.IPanic(and[0], obj.SyntaxPanic, "Comparison values are missing!")
				} else if index == len(and)-1 {
					fract.IPanic(and[len(and)-1], obj.SyntaxPanic, "Comparison values are missing!")
				}
				if !compare(*p.processValTokens(and[:index]), *p.processValTokens(and[index+1:]), operator) {
					return "false"
				}
			}
			return "true"
		}
		index, operator := findConditionOperator(or)
		// Operator is not found?
		if index == -1 {
			operator.Val = "=="
			if compare(*p.processValTokens(or), trueVal, operator) {
				return "true"
			}
			continue
		}
		// Operator is first or last?
		if index == 0 {
			fract.IPanic(or[0], obj.SyntaxPanic, "Comparison values are missing!")
		} else if index == len(or)-1 {
			fract.IPanic(or[len(or)-1], obj.SyntaxPanic, "Comparison values are missing!")
		}
		if compare(*p.processValTokens(or[:index]), *p.processValTokens(or[index+1:]), operator) {
			return "true"
		}
	}
	return "false"
}

// Returns string arithmetic compatible data.
func arithmetic(tk obj.Token, val oop.Val) string {
	result := val.String()
	switch val.Type {
	case oop.Func,
		oop.Package,
		oop.StructDef,
		oop.ClassDef,
		oop.ClassIns,
		oop.None:
		fract.IPanic(tk, obj.ArithmeticPanic, "\""+result+"\" is not compatible with arithmetic processes!")
	case oop.Map:
		fract.IPanic(tk, obj.ArithmeticPanic, "\"object.map\" is not compatible with arithmetic processes!")
	case oop.StructIns:
		fract.IPanic(tk, obj.ArithmeticPanic, "\"object.structins\" is not compatible with arithmetic processes!")
	}
	return result
}

// arithmeticProcess instance for solver.
type arithmeticProcess struct {
	left     []obj.Token
	leftVal  oop.Val
	right    []obj.Token
	rightVal oop.Val
	operator obj.Token
}

func (p arithmeticProcess) solve() oop.Val {
	val := oop.Val{Data: "0", Type: oop.Int}
	leftLen := p.leftVal.Len()
	rightLen := p.rightVal.Len()
	// String?
	if (leftLen != 0 && p.leftVal.Type == oop.String) || (rightLen != 0 && p.rightVal.Type == oop.String) {
		if p.leftVal.Type == p.rightVal.Type { // Both string?
			val.Type = oop.String
			switch p.operator.Val {
			case "+":
				val.Data = p.leftVal.String() + p.rightVal.String()
			case "-":
				if leftLen == 0 || rightLen == 0 {
					val.Data = ""
					break
				}
				if leftLen == 1 && rightLen > 1 {
					runeInt, _ := strconv.ParseInt(p.leftVal.String(), 10, 32)
					run := rune(runeInt)
					for _, r := range p.rightVal.String() {
						val.Data = val.String() + string(run-r)
					}
				} else if rightLen == 1 && leftLen > 1 {
					runeInt, _ := strconv.ParseInt(p.rightVal.String(), 10, 32)
					run := rune(runeInt)
					for _, r := range p.leftVal.String() {
						val.Data = val.String() + string(run-r)
					}
				} else {
					for i, r := range p.leftVal.String() {
						val.Data = val.String() + string(r-rune(p.rightVal.String()[i]))
					}
				}
			default:
				fract.IPanic(p.operator, obj.ArithmeticPanic, "This operator is not defined for string types!")
			}
			return val
		}

		val.Type = oop.String
		if p.rightVal.Type == oop.String {
			p.leftVal, p.rightVal = p.rightVal, p.leftVal
		}
		if p.rightVal.Type == oop.List {
			if rightLen == 0 {
				val.Data = p.leftVal.Data
				return val
			}
			if len(p.leftVal.String()) != rightLen && (len(p.leftVal.String()) != 1 && rightLen != 1) {
				fract.IPanic(p.right[0], obj.ArithmeticPanic, "List element count is not one or equals to first list!")
			}
			if strings.Contains(p.rightVal.String(), ".") {
				fract.IPanic(p.right[0], obj.ArithmeticPanic, "Only string and integer values can concatenate string values!")
			}
			runeInt, _ := strconv.ParseInt(p.rightVal.String(), 10, 64)
			run := rune(runeInt)
			var sb strings.Builder
			for _, r := range p.leftVal.String() {
				switch p.operator.Val {
				case "+":
					sb.WriteByte(byte(r + run))
				case "-":
					sb.WriteByte(byte(r - run))
				default:
					fract.IPanic(p.operator, obj.ArithmeticPanic, "This operator is not defined for string types!")
				}
			}
			val.Data = sb.String()
		} else {
			if p.rightVal.Type != oop.Int {
				fract.IPanic(p.right[0], obj.ArithmeticPanic, "Only string and integer values can concatenate string values!")
			}
			var str string
			runeInt, _ := strconv.ParseInt(p.rightVal.String(), 10, 64)
			run := byte(runeInt)
			for _, r := range p.leftVal.String() {
				switch p.operator.Val {
				case "+":
					str += string(byte(r) + run)
				case "-":
					str += string(byte(r) - run)
				default:
					fract.IPanic(p.operator, obj.ArithmeticPanic, "This operator is not defined for string types!")
				}
			}
			val.Data = str
		}
		return val
	}

	if p.leftVal.Type == oop.List && p.rightVal.Type == oop.List {
		val.Type = oop.List
		if leftLen == 0 {
			val.Data = p.rightVal.Data
			return val
		} else if rightLen == 0 {
			val.Data = p.leftVal.Data
			return val
		}
		if leftLen != rightLen && leftLen != 1 && rightLen != 1 {
			fract.IPanic(p.right[0], obj.ArithmeticPanic, "List element count is not one or equals to first list!")
		}
		if leftLen == 1 || rightLen == 1 {
			left, right := p.leftVal, p.rightVal
			if left.Len() != 1 {
				left, right = right, left
			}
			arith := str.Conv(arithmetic(p.operator, left.Data.(*oop.ListModel).Elems[0]))
			for i, elem := range right.Data.(*oop.ListModel).Elems {
				if elem.Type == oop.List {
					right.Data.(*oop.ListModel).Elems[i] = readyData(p, oop.Val{
						Data: arithmeticProcess{
							left:     p.left,
							leftVal:  right,
							right:    p.right,
							rightVal: elem,
							operator: p.operator,
						}.solve().Data,
						Type: oop.List,
					})
				} else {
					right.Data.(*oop.ListModel).Elems[i] = readyData(p, oop.Val{
						Data: fmt.Sprintf(fract.FloatFormat, solveArithmeticProcess(p.operator, arith, str.Conv(arithmetic(p.operator, elem)))),
						Type: oop.Int,
					})
				}
			}
			val.Data = right.Data
		} else {
			for i, elem := range p.leftVal.Data.(*oop.ListModel).Elems {
				right := p.rightVal.Data.(*oop.ListModel).Elems[i]
				if elem.Type == oop.List || right.Type == oop.List {
					proc := arithmeticProcess{left: p.left, right: p.right, operator: p.operator}
					if elem.Type == oop.List {
						proc.leftVal = oop.Val{Data: elem.Data, Type: oop.List}
					} else {
						proc.leftVal = elem
					}
					if right.Type == oop.List {
						proc.rightVal = oop.Val{Data: right.Data, Type: oop.List}
					} else {
						proc.rightVal = right
					}
					p.leftVal.Data.(*oop.ListModel).Elems[i] = readyData(p, oop.Val{
						Data: proc.solve().Data,
						Type: oop.List,
					})
				} else {
					p.leftVal.Data.(*oop.ListModel).Elems[i] = readyData(p, oop.Val{
						Data: fmt.Sprintf(fract.FloatFormat, solveArithmeticProcess(p.operator, str.Conv(arithmetic(p.operator, elem)), str.Conv(right.String()))),
						Type: oop.Int,
					})
				}
			}
			val.Data = p.leftVal.Data
		}
	} else if p.leftVal.Type == oop.List || p.rightVal.Type == oop.List {
		val.Type = oop.List
		if p.leftVal.Type == oop.List && leftLen == 0 {
			val.Data = p.rightVal.Data
			return val
		} else if p.rightVal.Type == oop.List && rightLen == 0 {
			val.Data = p.leftVal.Data
			return val
		}
		left, right := p.leftVal, p.rightVal
		if left.Type != oop.List {
			left, right = right, left
		}
		arith := str.Conv(arithmetic(p.operator, right))
		for i, elem := range left.Data.(*oop.ListModel).Elems {
			if elem.Type == oop.List {
				left.Data.(*oop.ListModel).Elems[i] = readyData(p, arithmeticProcess{
					left:     p.left,
					leftVal:  right,
					right:    p.right,
					rightVal: elem,
					operator: p.operator,
				}.solve())
			} else {
				left.Data.(*oop.ListModel).Elems[i] = readyData(p, oop.Val{
					Data: fmt.Sprintf(fract.FloatFormat, solveArithmeticProcess(p.operator, str.Conv(arithmetic(p.operator, elem)), arith)),
					Type: oop.Int,
				})
			}
		}
		val = left
	} else {
		val = readyData(p, oop.Val{
			Data: fmt.Sprintf(fract.FloatFormat, solveArithmeticProcess(p.operator, str.Conv(arithmetic(p.operator, p.leftVal)), str.Conv(arithmetic(p.operator, p.rightVal)))),
			Type: oop.Int,
		})
	}
	return val
}

func solveArithmeticProcess(operator obj.Token, left, right float64) float64 {
	var result float64
	switch operator.Val {
	case "+": // Addition.
		result = left + right
	case "-": // Subtraction.
		result = left - right
	case "*": // Multiply.
		result = left * right
	case "/", "//": // Division.
		if left == 0 || right == 0 {
			fract.Panic(operator, obj.DivideByZeroPanic, "Divide by zero!")
		}
		result = left / right
	case "|": // Binary or.
		result = float64(int(left) | int(right))
	case "&": // Binary and.
		result = float64(int(left) & int(right))
	case "^": // Bitwise exclusive or.
		result = float64(int(left) ^ int(right))
	case "**": // Exponentiation.
		result = math.Pow(left, right)
	case "%": // Mod.
		result = math.Mod(left, right)
	case "<<": // Left shift.
		if right < 0 {
			fract.IPanic(operator, obj.ArithmeticPanic, "Shifter is cannot should be negative!")
		}
		result = float64(int(left) << int(right))
	case ">>": // Right shift.
		if right < 0 {
			fract.IPanic(operator, obj.ArithmeticPanic, "Shifter is cannot should be negative!")
		}
		result = float64(int(left) >> int(right))
	default:
		fract.IPanic(operator, obj.SyntaxPanic, "Operator is invalid!")
	}
	return result
}

// Check data and set ready.
func readyData(process arithmeticProcess, val oop.Val) oop.Val {
	if process.leftVal.Type == oop.String || process.rightVal.Type == oop.String {
		val.Type = oop.String
	} else if process.operator.Val == "/" || process.leftVal.Type == oop.Float || process.rightVal.Type == oop.Float {
		val.Type = oop.Float
		return val
	}
	return val
}

// Select elements of enumerable object.
func (p *Parser) selectEnumerable(mut bool, v oop.Val, tk obj.Token, s interface{}) *oop.Val {
	var result oop.Val
	switch v.Type {
	case oop.List:
		index := s.([]int)
		if len(index) == 1 {
			val := v.Data.(*oop.ListModel).Elems[index[0]]
			if !val.Mut && !mut { //! Immutability.
				val = val.Immut()
			}
			val.Mut = val.Mut || mut
			return &val
		}
		list := oop.NewListModel()
		for _, pos := range index {
			list.PushBack(v.Data.(*oop.ListModel).Elems[pos])
		}
		result = oop.Val{Data: list, Type: oop.List}
	case oop.Map:
		m := v.Data.(oop.MapModel).Map
		switch t := s.(type) {
		case oop.ListModel:
			resultMap := oop.NewMapModel()
			for _, key := range t.Elems {
				val, ok := m[key]
				if !ok {
					fract.IPanic(tk, obj.ValuePanic, "Key is not exists!")
				}
				resultMap.Map[key] = val
			}
			result = oop.Val{Data: resultMap, Type: oop.Map}
		case oop.Val:
			val, ok := m[t]
			if !ok {
				fract.IPanic(tk, obj.ValuePanic, "Key is not exists!")
			}
			return &val
		}
	case oop.String:
		result = oop.Val{Data: oop.NewStringModel(""), Type: oop.String}
		for _, i := range s.([]int) {
			result.Data = result.String() + string(v.String()[i])
		}
	}
	return &result
}

type valuePartInfo struct {
	tokens []obj.Token
	mut    bool // Force to mutability.
}

func (p *Parser) processNameValue(mut bool, tk obj.Token) *oop.Val {
	var result *oop.Val
	defIndex, defType := p.defByName(tk.Val)
	if defIndex == -1 {
		if tk.Val == "this" {
			fract.IPanic(tk, obj.NamePanic, `"this" keyword is cannot used this scope!`)
		}
		fract.IPanic(tk, obj.NamePanic, "Name is not defined: "+tk.Val)
	}
	switch defType {
	case 'f': // Function.
		result = &oop.Val{Data: p.defs.Funcs[defIndex], Type: oop.Func}
	case 'p': // Package.
		result = &oop.Val{Data: p.packages[defIndex], Type: oop.Package}
	case 'v': // Value.
		val := p.defs.Vars[defIndex]
		result = &val.Val
		if !val.Val.Mut && !mut { //! Immutability.
			*result = val.Val.Immut()
		}
		result.Mut = val.Val.Mut || mut
	}
	return result
}

func (p *Parser) processValuePart(part valuePartInfo) *oop.Val {
	var result *oop.Val
	if part.tokens[0].Type == fract.Var && part.tokens[0].Val == "mut" {
		if len(part.tokens) == 1 {
			fract.IPanic(part.tokens[0], obj.SyntaxPanic, "Value is not given!")
		}
		part.mut = true
		part.tokens = part.tokens[1:]
		result = p.processValuePart(part)
		goto end
	}
	// Single oop.
	if tk := part.tokens[0]; len(part.tokens) == 1 {
		if tk.Val[0] == '\'' || tk.Val[0] == '"' {
			result = &oop.Val{Data: oop.NewStringModel(tk.Val[1 : len(tk.Val)-1]), Type: oop.String}
			goto end
		} else if tk.Val == "true" || tk.Val == "false" {
			result = &oop.Val{Data: tk.Val, Type: oop.Bool}
			goto end
		} else if tk.Val == "none" {
			result = &oop.Val{Data: tk.Val, Type: oop.None}
			goto end
		} else if tk.Type == fract.Value {
			if strings.Contains(tk.Val, ".") || strings.ContainsAny(tk.Val, "eE") {
				tk.Type = oop.Float
			} else {
				tk.Type = oop.Int
			}
			if tk.Val != "NaN" {
				prs, _ := new(big.Float).SetString(tk.Val)
				val, _ := prs.Float64()
				tk.Val = fmt.Sprint(val)
			}
			result = &oop.Val{Data: tk.Val, Type: tk.Type}
			goto end
		} else {
			if tk.Type != fract.Name {
				fract.IPanic(tk, obj.ValuePanic, "Invalid value!")
			}
		}
	}
	switch j, tk := len(part.tokens)-1, part.tokens[len(part.tokens)-1]; tk.Type {
	case fract.Name:
		if j > 0 {
			j--
			if j == 0 || part.tokens[j].Type != fract.Dot {
				fract.IPanic(part.tokens[j], obj.SyntaxPanic, "Invalid syntax!")
			}
			nameTk := part.tokens[j+1]
			valTk := part.tokens[j]
			part.tokens = part.tokens[:j]
			part.mut = true
			val := p.processValuePart(part)
			part.mut = false
			switch val.Type {
			case oop.Package:
				impInf := val.Data.(importInfo)
				checkPublic(nil, nameTk)
				result = impInf.src.processNameValue(part.mut, nameTk)
				goto end
			case oop.StructIns:
				ins := val.Data.(oop.StructInstance)
				checkPublic(ins.File, tk)
				i := ins.Fields.VarIndexByName(nameTk.Val)
				if i == -1 {
					fract.IPanic(nameTk, obj.NamePanic, "Name is not defined: "+nameTk.Val)
				}
				result = &ins.Fields.Vars[i].Val
				goto end
			case oop.Map:
				m := val.Data.(oop.MapModel)
				i := m.Defs.FuncIndexByName(nameTk.Val)
				if i == -1 {
					fract.IPanic(nameTk, obj.NamePanic, "Name is not defined: "+nameTk.Val)
				}
				result = &oop.Val{Data: m.Defs.Funcs[i], Type: oop.Func}
				goto end
			case oop.ClassIns:
				ins := val.Data.(oop.ClassInstance)
				checkPublic(ins.File, tk)
				defIndex, defType := ins.Defs.DefByName(nameTk.Val)
				if defIndex == -1 {
					fract.IPanic(nameTk, obj.NamePanic, "Name is not defined: "+nameTk.Val)
				}
				switch defType {
				case 'f': // Function.
					result = &oop.Val{Data: ins.Defs.Funcs[defIndex], Type: oop.Func}
				case 'v': // Value.
					result = &ins.Defs.Vars[defIndex].Val
					if !result.Mut && !part.mut { //! Immutability.
						*result = result.Immut()
					}
					result.Mut = result.Mut || part.mut
				}
				goto end
			case oop.List:
				list := val.Data.(*oop.ListModel)
				i := list.Defs.FuncIndexByName(nameTk.Val)
				if i == -1 {
					fract.IPanic(nameTk, obj.NamePanic, "Name is not defined: "+nameTk.Val)
				}
				result = &oop.Val{Data: list.Defs.Funcs[i], Type: oop.Func}
				goto end
			case oop.String:
				str := val.Data.(oop.StringModel)
				i := str.Defs.FuncIndexByName(nameTk.Val)
				if i == -1 {
					fract.IPanic(nameTk, obj.NamePanic, "Name is not defined: "+nameTk.Val)
				}
				result = &oop.Val{Data: str.Defs.Funcs[i], Type: oop.Func}
				goto end
			default:
				fract.IPanic(valTk, obj.ValuePanic, "Object is not support sub fields!")
			}
		}
		result = p.processNameValue(part.mut, tk)
		goto end
	case fract.Brace:
		braceCount := 0
		switch tk.Val {
		case ")":
			var valTokens []obj.Token
			for ; j >= 0; j-- {
				tk := part.tokens[j]
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
				valTokens = part.tokens[:j]
				break
			}
			if len(valTokens) == 0 && braceCount == 0 {
				tk, part.tokens = part.tokens[0], part.tokens[1:len(part.tokens)-1]
				if len(part.tokens) == 0 {
					fract.IPanic(tk, obj.SyntaxPanic, "Invalid syntax!")
				}
				result = p.processValue(part.tokens, part.mut)
				goto end
			}
			val := p.processValuePart(valuePartInfo{tokens: valTokens, mut: part.mut})
			switch val.Type {
			case oop.Func: // Function call.
				result = p.funcCallModel(val.Data.(*oop.Fn), part.tokens[len(valTokens):]).Call()
			case oop.StructDef:
				s := val.Data.(oop.Struct)
				result = &oop.Val{
					Data: s.CallConstructor(p.funcCallModel(s.Constructor, part.tokens[len(valTokens):]).args),
					Type: oop.StructIns,
				}
			case oop.ClassDef:
				class := val.Data.(oop.Class)
				result = &oop.Val{
					Data: class.CallConstructor(p.funcCallModel(class.Constructor, part.tokens[len(valTokens):])),
					Type: oop.ClassIns,
				}
			default:
				fract.IPanic(part.tokens[len(valTokens)], obj.ValuePanic, "Invalid syntax!")
			}
			goto end
		case "]":
			var valTokens []obj.Token
			for ; j >= 0; j-- {
				tk := part.tokens[j]
				if tk.Type != fract.Brace {
					continue
				}
				switch tk.Val {
				case "]":
					braceCount++
				case "[":
					braceCount--
				}
				if braceCount > 0 {
					continue
				}
				valTokens = part.tokens[:j]
				break
			}
			if len(valTokens) == 0 && braceCount == 0 {
				result = p.processEnumerableValue(part.tokens)
				goto end
			}
			val := p.processValuePart(valuePartInfo{mut: part.mut, tokens: valTokens})
			if !val.IsEnum() {
				fract.IPanic(valTokens[0], obj.ValuePanic, "Index accessor is cannot used with not enumerable values!")
			}
			result = p.selectEnumerable(part.mut, *val, tk, enumerableSelections(*val, *p.processValTokens(part.tokens[len(valTokens)+1 : len(part.tokens)-1]), tk))
			goto end
		case "}":
			var valTokens []obj.Token
			for ; j >= 0; j-- {
				tk := part.tokens[j]
				if tk.Type != fract.Brace {
					continue
				}
				switch tk.Val {
				case "}":
					braceCount++
				case "{":
					braceCount--
				}
				if braceCount > 0 {
					continue
				}
				valTokens = part.tokens[:j]
				break
			}
			valTokensLen := len(valTokens)
			if valTokensLen == 0 && braceCount == 0 {
				result = p.processEnumerableValue(part.tokens)
				goto end
			} else if valTokensLen > 1 && (valTokens[1].Type != fract.Brace || valTokens[1].Val != "(") {
				fract.IPanic(valTokens[1], obj.SyntaxPanic, "Invalid syntax!")
			} else if valTokensLen > 1 && (valTokens[valTokensLen-1].Type != fract.Brace || valTokens[valTokensLen-1].Val != ")") {
				fract.IPanic(valTokens[valTokensLen-1], obj.SyntaxPanic, "Invalid syntax!")
			}
			switch valTokens[0].Type {
			case fract.Fn:
				fn := &oop.Fn{
					Name:   "anonymous",
					Src:    p,
					Tokens: p.getBlock(part.tokens[len(valTokens):]),
				}
				if fn.Tokens == nil {
					fn.Tokens = [][]obj.Token{}
				}
				if valTokensLen > 1 {
					valTokens = valTokens[1:]
					valTokens = decomposeBrace(&valTokens)
					p.setParams(fn, &valTokens)
				}
				result = &oop.Val{Data: fn, Type: oop.Func}
			case fract.Struct:
				result = p.buildStruct("anonymous", part.tokens[1:])
			default:
				fract.IPanic(valTokens[0], obj.SyntaxPanic, "Invalid syntax!")
			}
			valTokens = nil
			goto end
		}
	}
	fract.IPanic(part.tokens[0], obj.ValuePanic, "Invalid value!")
end:
	result.Mut = part.mut
	return result
}

func (p *Parser) processListValue(tokens []obj.Token) *oop.Val {
	var braceCount int
	comma := 1
	list := oop.NewListModel()
	for j := 1; j < len(tokens)-1; j++ {
		switch typ := tokens[j]; typ.Type {
		case fract.Brace:
			switch typ.Val {
			case "{", "[", "(":
				braceCount++
			default:
				braceCount--
			}
		case fract.Comma:
			if braceCount != 0 {
				break
			}
			if comma-j == 0 {
				fract.IPanic(typ, obj.SyntaxPanic, "Value is not given!")
			}
			list.PushBack(*p.processValTokens(tokens[comma:j]))
			comma = j + 1
		}
	}
	if len := len(tokens); comma < len-1 {
		list.PushBack(*p.processValTokens(tokens[comma : len-1]))
	}
	return &oop.Val{Data: list, Type: oop.List}
}

func (p *Parser) processMapValue(tokens []obj.Token) *oop.Val {
	var braceCount int
	m := oop.NewMapModel()
	comma := 1
	for j := 1; j < len(tokens)-1; j++ {
		switch typ := tokens[j]; typ.Type {
		case fract.Brace:
			switch typ.Val {
			case "{", "[", "(":
				braceCount++
			default:
				braceCount--
			}
		case fract.Comma:
			if braceCount != 0 {
				break
			}
			if comma-j == 0 {
				fract.IPanic(typ, obj.SyntaxPanic, "Value is not given!")
			}
			modelTokens := tokens[comma:j]
			var (
				i              int
				modelTokensLen int = len(modelTokens)
				tk             obj.Token
			)
			for i, tk = range modelTokens {
				switch tk.Type {
				case fract.Brace:
					switch tk.Val {
					case "{", "[", "(":
						braceCount++
					default:
						braceCount--
					}
				case fract.Colon:
					if braceCount != 0 {
						break
					}
					if i+1 >= modelTokensLen {
						fract.IPanic(tk, obj.SyntaxPanic, "Value is not given!")
					}
					key := *p.processValTokens(modelTokens[:i])
					_, ok := m.Map[key]
					if ok {
						fract.IPanic(tk, obj.ValuePanic, "Key is already defined!")
					}
					m.Map[key] = *p.processValTokens(modelTokens[i+1:])
					comma = j + 1
					modelTokens = nil
				}
			}
			if modelTokens != nil {
				fract.IPanic(modelTokens[modelTokensLen-1], obj.SyntaxPanic, "Value identifier is not found!")
			}
		}
	}
	if comma < len(tokens)-1 {
		lastTokens := tokens[comma : len(tokens)-1]
		i := -1
		lenLastTokens := len(lastTokens)
		for j, tk := range lastTokens {
			switch tk.Type {
			case fract.Brace:
				switch tk.Val {
				case "{", "[", "(":
					braceCount++
				default:
					braceCount--
				}
			case fract.Colon:
				if braceCount != 0 {
					break
				}
				i = j
			}
			if i != -1 {
				break
			}
		}
		if i+1 >= lenLastTokens {
			fract.IPanic(lastTokens[i], obj.SyntaxPanic, "Value is not given!")
		}
		key := *p.processValTokens(lastTokens[:i])
		_, ok := m.Map[key]
		if ok {
			fract.IPanic(lastTokens[i], obj.ValuePanic, "Key is already defined!")
		}
		m.Map[key] = *p.processValTokens(lastTokens[i+1:])
		lastTokens = nil
	}
	return &oop.Val{Data: m, Type: oop.Map}
}

func (p *Parser) processListComprehension(tokens []obj.Token) *oop.Val {
	var (
		selectTokens []obj.Token
		loopTokens   []obj.Token
		filterTokens []obj.Token
		braceCount   int
	)
	for i, tk := range tokens {
		if tk.Type == fract.Brace {
			switch tk.Val {
			case "{", "[", "(":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 1 {
			continue
		}
		if tk.Type == fract.Loop {
			selectTokens = tokens[1:i]
		} else if tk.Type == fract.Comma {
			loopTokens = tokens[len(selectTokens)+1 : i]
			filterTokens = tokens[i+1 : len(tokens)-1]
			if len(filterTokens) == 0 {
				filterTokens = nil
			}
			break
		}
	}
	if loopTokens == nil {
		loopTokens = tokens[len(selectTokens)+1 : len(tokens)-1]
	}
	if len(loopTokens) < 2 {
		fract.IPanic(loopTokens[0], obj.SyntaxPanic, "Variable name is not given!")
	}
	nameTk := loopTokens[1]
	// Name is not name?
	if nameTk.Type != fract.Name {
		fract.IPanic(nameTk, obj.SyntaxPanic, "This is not a valid name!")
	}
	if line := p.defIndexByName(nameTk.Val); line != -1 {
		fract.IPanic(nameTk, obj.NamePanic, "\""+nameTk.Val+"\" is already defined at line: "+fmt.Sprint(line))
	}
	if lenLoopTokens := len(loopTokens); lenLoopTokens < 3 {
		tk := tokens[0]
		fract.IPanicC(tk.File, tk.Line, loopTokens[1].Column+len(loopTokens[1].Val), obj.SyntaxPanic, "Value is not given!")
	} else if tk := loopTokens[2]; tk.Type != fract.In && (tk.Type != fract.Operator || tk.Val != ":=") {
		fract.IPanic(loopTokens[2], obj.SyntaxPanic, "Invalid syntax!")
	} else if lenLoopTokens < 4 {
		fract.IPanic(loopTokens[2], obj.SyntaxPanic, "Value is not given!")
	}
	loopTokens = loopTokens[3:]
	varVal := *p.processValTokens(loopTokens)
	if !varVal.IsEnum() {
		fract.IPanic(loopTokens[0], obj.ValuePanic, "Foreach loop must defined enumerable value!")
	}
	if nameTk.Val == "_" {
		nameTk.Val = ""
	} else if !isValidName(nameTk.Val) {
		fract.IPanic(nameTk, obj.NamePanic, "Invalid name!")
	}
	p.defs.Vars = append(p.defs.Vars, oop.Var{Name: nameTk.Val})
	elem := &p.defs.Vars[len(p.defs.Vars)-1]
	// Interpret block.
	list := oop.NewListModel()
	l := loop{val: varVal}
	l.run(func() {
		elem.Val = l.b
		if filterTokens == nil || p.prococessCondition(filterTokens) == "true" {
			list.PushBack(*p.processValTokens(selectTokens))
		}
	})
	// Remove variables.
	elem = nil
	p.defs.Vars = p.defs.Vars[:len(p.defs.Vars)-1]
	return &oop.Val{Data: list, Type: oop.List}
}

func (p *Parser) processEnumerableValue(tokens []obj.Token) *oop.Val {
	var (
		ListComprehension bool
		braceCount        int
	)
	for _, t := range tokens {
		if t.Type == fract.Brace {
			switch t.Val {
			case "{", "[", "(":
				braceCount++
			default:
				braceCount--
			}
		}
		if braceCount > 1 {
			continue
		}
		if t.Type == fract.Comma {
			break
		} else if !ListComprehension && t.Type == fract.Loop {
			if tokens[0].Val != "[" {
				fract.IPanic(tokens[0], obj.SyntaxPanic, "Invalid syntax!")
			}
			ListComprehension = true
			break
		}
	}
	if ListComprehension {
		return p.processListComprehension(tokens)
	} else if tokens[0].Val == "{" {
		return p.processMapValue(tokens)
	}
	return p.processListValue(tokens)
}

func (p *Parser) processValue(tks []obj.Token, mut bool) *oop.Val {
	// Is conditional expression?
	if j, _ := findConditionOperator(tks); j != -1 {
		return &oop.Val{Data: p.prococessCondition(tks), Type: oop.Bool}
	}
	processes := arithmeticProcesses(tks)
	part := valuePartInfo{mut: mut}
	if len(processes) == 1 {
		part.tokens = processes[0]
		return p.processValuePart(part)
	}
	var result oop.Val
	var process arithmeticProcess
	j := nextOperator(processes)
	for j != -1 {
		if j == 0 {
			if len(processes) == 1 {
				break
			}
			process.leftVal = result
			process.operator = processes[j][0]
			process.right = processes[j+1]
			part.tokens = process.right
			process.rightVal = *p.processValuePart(part)
			if process.rightVal.Type == fract.NA {
				fract.IPanic(process.left[0], obj.ValuePanic, "Value is not given!")
			}
			result = process.solve()
			processes = processes[2:]
			j = nextOperator(processes)
			continue
		}
		process.left = processes[j-1]
		part.tokens = process.left
		process.leftVal = *p.processValuePart(part)
		if process.leftVal.Type == fract.NA {
			fract.IPanic(process.left[0], obj.ValuePanic, "Value is not given!")
		}
		process.operator = processes[j][0]
		process.right = processes[j+1]
		part.tokens = process.right
		process.rightVal = *p.processValuePart(part)
		if process.rightVal.Type == fract.NA {
			fract.IPanic(process.right[0], obj.ValuePanic, "Value is not given!")
		}
		val := process.solve()
		if result.Data != nil {
			process.operator.Val = "+"
			process.right = processes[j+1]
			process.leftVal = result
			process.rightVal = val
			result = process.solve()
		} else {
			result = val
		}
		// Remove computed processes.
		processes = append(processes[:j-1], processes[j+2:]...)
		// Find next operator.
		j = nextOperator(processes)
	}
	processes = nil
	process.left = nil
	process.right = nil
	return &result
}

func (p *Parser) processValTokens(tks []obj.Token) *oop.Val {
	return p.processValue(tks, false)
}
