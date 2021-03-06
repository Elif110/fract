/*
	processValue Function
*/

package interpreter

import (
	"strings"

	"../fract"
	"../fract/arithmetic"
	"../fract/name"
	"../grammar"
	"../objects"
	"../parser"
	"../utilities/vector"
)

// __processValue Process value.
// first This is first value.
// token Token to process.
// operations All operations.
// index Index of token.
func (i *Interpreter) _processValue(first bool, operation *objects.ArithmeticProcess,
	operations *vector.Vector, index int) int {
	token := operation.First
	if !first {
		token = operation.Second
	}

	if token.Type == fract.TypeName {
		if index < len(operations.Vals)-1 {
			next := operations.Vals[index+1].(objects.Token)
			// Array?
			if next.Type == fract.TypeBrace {
				if next.Value == grammar.TokenLBracket {
					vindex := name.VarIndexByName(i.vars, token.Value)
					if vindex == -1 {
						fract.Error(token, "Name is not defined!: "+token.Value)
					}

					// Find close bracket.
					cindex := index + 1
					bracketCount := 1
					for ; cindex < len(operations.Vals); cindex++ {
						current := operations.Vals[cindex].(objects.Token)
						if current.Type == fract.TypeBrace {
							if current.Value == grammar.TokenLBracket {
								bracketCount++
							} else if current.Value == grammar.TokenRBracket {
								bracketCount--
							}
						}

						if bracketCount == 0 {
							break
						}
					}

					valueList := operations.Sublist(index+2, cindex-index-3)
					// Index value is empty?
					if len(valueList.Vals) == 0 {
						fract.Error(token, "Index is not defined!")
					}

					value := i.processValue(valueList)
					if value.Array {
						fract.Error(operations.Vals[index].(objects.Token),
							"Arrays is not used in index access!")
					} else if arithmetic.IsFloatValue(value.Content[0]) {
						fract.Error(operations.Vals[index].(objects.Token),
							"Float values is not used in index access!")
					}
					position, err := arithmetic.ToInt64(value.Content[0])
					if err != nil {
						fract.Error(operations.Vals[index].(objects.Token),
							"Value out of range!")
					}

					variable := i.vars.Vals[vindex].(objects.Variable)

					if !variable.Value.Array {
						fract.Error(operations.Vals[index].(objects.Token),
							"Index accessor is cannot used with non-array variables!")
					}

					if position < 0 || position >= int64(len(variable.Value.Content)) {
						fract.Error(operations.Vals[index].(objects.Token),
							"Index is out of range!")
					}
					operations.RemoveRange(index+1, cindex-index-1)
					if first {
						operation.FirstV.Content = []string{variable.Value.Content[position]}
						operation.FirstV.Array = false
					} else {
						operation.SecondV.Content = []string{variable.Value.Content[position]}
						operation.SecondV.Array = false
					}
					return 0
				} else if next.Value == grammar.TokenLParenthes { // Function?
					// Find close parentheses.
					cindex := index + 1
					bracketCount := 1
					for ; cindex < len(operations.Vals); cindex++ {
						current := operations.Vals[cindex].(objects.Token)
						if current.Type == fract.TypeBrace {
							if current.Value == grammar.TokenLParenthes {
								bracketCount++
							} else if current.Value == grammar.TokenRParenthes {
								bracketCount--
							}
						}

						if bracketCount == 0 {
							break
						}
					}
					value := i.processFunctionCall(operations.Sublist(index, cindex-index))
					if !operation.FirstV.Array && value.Content == nil {
						fract.Error(token, "Function is not return any value!")
					}
					operations.RemoveRange(index+1, cindex-index-1)
					if first {
						operation.FirstV = value
					} else {
						operation.SecondV = value
					}
					return 0
				}
			}
		}

		vindex := name.VarIndexByName(i.vars, token.Value)
		if vindex == -1 {
			fract.Error(token, "Name is not defined!: "+token.Value)
		}

		variable := i.vars.Vals[vindex].(objects.Variable)

		if first {
			operation.FirstV = variable.Value
		} else {
			operation.SecondV = variable.Value
		}
		return 0
	} else if token.Type == fract.TypeBrace && token.Value == grammar.TokenRBracket {
		// Find open bracket.
		bracketCount := 1
		oindex := index - 1
		for ; oindex >= 0; oindex-- {
			current := operations.Vals[oindex].(objects.Token)
			if current.Type == fract.TypeBrace {
				if current.Value == grammar.TokenRBracket {
					bracketCount++
				} else if current.Value == grammar.TokenLBracket {
					bracketCount--
				}
			}

			if bracketCount == 0 {
				break
			}
		}

		// Finished?
		if oindex == 0 {
			if first {
				operation.FirstV.Array = true
				operation.FirstV.Content = i.processArrayValue(
					operations.Sublist(oindex, index-oindex+1)).Content
			} else {
				operation.SecondV.Array = true
				operation.SecondV.Content = i.processArrayValue(
					operations.Sublist(oindex, index-oindex+1)).Content
			}
			operations.RemoveRange(oindex, index-oindex)
			return index - oindex
		}

		endToken := operations.Vals[oindex-1].(objects.Token)
		vindex := name.VarIndexByName(i.vars, endToken.Value)
		if vindex == -1 {
			fract.Error(endToken, "Name is not defined!: "+endToken.Value)
		}
		valueList := operations.Sublist(oindex+1, index-oindex-1)
		// Index value is empty?
		if len(valueList.Vals) == 0 {
			fract.Error(endToken, "Index is not defined!")
		}

		value := i.processValue(valueList)
		if value.Array {
			fract.Error(operations.Vals[index].(objects.Token),
				"Arrays is not used in index access!")
		} else if arithmetic.IsFloatValue(value.Content[0]) {
			fract.Error(operations.Vals[index].(objects.Token),
				"Float values is not used in index access!")
		}

		position, err := arithmetic.ToInt64(value.Content[0])
		if err != nil {
			fract.Error(operations.Vals[oindex].(objects.Token), "Value out of range!")
		}

		variable := i.vars.Vals[vindex].(objects.Variable)

		if !variable.Value.Array {
			fract.Error(operations.Vals[oindex].(objects.Token),
				"Index accessor is cannot used with non-array variables!")
		}

		if position < 0 || position >= int64(len(variable.Value.Content)) {
			fract.Error(operations.Vals[oindex].(objects.Token), "Index is out of range!")
		}
		operations.RemoveRange(oindex-1, index-oindex+1)

		if first {
			operation.FirstV.Content = []string{variable.Value.Content[position]}
			operation.FirstV.Array = false
		} else {
			operation.SecondV.Content = []string{variable.Value.Content[position]}
			operation.SecondV.Array = false
		}

		return index - oindex + 1
	} else if token.Type == fract.TypeBrace && token.Value == grammar.TokenLBracket {
		// Array constructor.
		cindex := index + 1
		bracketCount := 1
		for ; cindex < len(operations.Vals); cindex++ {
			current := operations.Vals[cindex].(objects.Token)
			if current.Type == fract.TypeBrace {
				if current.Value == grammar.TokenLBracket {
					bracketCount++
				} else if current.Value == grammar.TokenRBracket {
					bracketCount--
				}
			}

			if bracketCount == 0 {
				break
			}
		}

		if first {
			operation.FirstV.Array = true
			operation.FirstV.Content = i.processArrayValue(
				operations.Sublist(index, cindex-index+1)).Content
		} else {
			operation.SecondV.Array = true
			operation.SecondV.Content = i.processArrayValue(
				operations.Sublist(index, cindex-index+1)).Content
		}
		operations.RemoveRange(index+1, cindex-index-1)
		return 0
	} else if token.Type == fract.TypeBrace && token.Value == grammar.TokenLBrace {
		// Array initializer.

		// Find close brace.
		cindex := index + 1
		braceCount := 1
		for ; cindex < len(operations.Vals); cindex++ {
			current := operations.Vals[cindex].(objects.Token)
			if current.Type == fract.TypeBrace {
				if current.Value == grammar.TokenLBrace {
					fract.Error(current, "Arrays is cannot take array value as element!")
					braceCount++
				} else if current.Value == grammar.TokenRBrace {
					braceCount--
				}
			}

			if braceCount == 0 {
				break
			}
		}

		value := i.processArrayValue(operations.Sublist(index, cindex-index+1))
		if first {
			operation.FirstV = value
		} else {
			operation.SecondV = value
		}
		operations.RemoveRange(index+1, cindex-index-1)
		return 0
	} else if token.Type == fract.TypeBrace && token.Value == grammar.TokenRBrace {
		// Array initializer.

		// Find open brace.
		braceCount := 1
		oindex := index - 1
		nestedArray := false
		for ; oindex >= 0; oindex-- {
			current := operations.Vals[oindex].(objects.Token)
			if current.Type == fract.TypeBrace {
				if current.Value == grammar.TokenRBrace {
					braceCount++
					nestedArray = true
				} else if current.Value == grammar.TokenLBrace {
					braceCount--
					if nestedArray {
						fract.Error(current, "Arrays is cannot take array value as element!")
					}
				}
			}

			if braceCount == 0 {
				break
			}
		}

		value := i.processArrayValue(operations.Sublist(oindex, index-oindex+1))
		if first {
			operation.FirstV = value
		} else {
			operation.SecondV = value
		}
		operations.RemoveRange(oindex, index-oindex)
		return index - oindex
	} else if token.Type == fract.TypeBrace && token.Value == grammar.TokenRParenthes {
		// Function.

		// Find open parentheses.
		bracketCount := 1
		oindex := index - 1
		for ; oindex >= 0; oindex-- {
			current := operations.Vals[oindex].(objects.Token)
			if current.Type == fract.TypeBrace {
				if current.Value == grammar.TokenRBracket {
					bracketCount++
				} else if current.Value == grammar.TokenLBracket {
					bracketCount--
				}
			}

			if bracketCount == 0 {
				break
			}
		}
		oindex++
		value := i.processFunctionCall(operations.Sublist(oindex, index-oindex+1))
		if value.Content == nil {
			fract.Error(operations.Vals[oindex].(objects.Token),
				"Function is not return any value!")
		}
		if first {
			operation.FirstV = value
		} else {
			operation.SecondV = value
		}
		operations.RemoveRange(oindex, index-oindex)
		return index - oindex
	}

	//
	// Single value.
	//

	if !strings.HasPrefix(token.Value, grammar.TokenQuote) &&
		!strings.HasPrefix(token.Value, grammar.TokenDoubleQuote) {
		_, err := arithmetic.ToFloat64(token.Value)
		if err != nil {
			fract.Error(token, "Value out of range!")
		}
	}

	// Boolean check.
	if token.Type == fract.TypeBooleanTrue {
		token.Value = "1"
	} else if token.Type == fract.TypeBooleanFalse {
		token.Value = "0"
	}

	if first {
		operation.FirstV.Array = false
		if strings.HasPrefix(token.Value, grammar.TokenQuote) ||
			strings.HasPrefix(token.Value, grammar.TokenDoubleQuote) { // String?
			operation.FirstV.Charray = true
			operation.FirstV.Array = true
			for index := 1; index < len(token.Value)-1; index++ {
				operation.FirstV.Content = append(
					operation.FirstV.Content, arithmetic.IntToString(token.Value[index]))
			}
		} else {
			operation.FirstV.Content = []string{token.Value}
		}
	} else {
		operation.SecondV.Array = false
		if strings.HasPrefix(token.Value, grammar.TokenQuote) ||
			strings.HasPrefix(token.Value, grammar.TokenDoubleQuote) { // String?
			operation.SecondV.Charray = true
			operation.SecondV.Array = true
			for index := 1; index < len(token.Value)-1; index++ {
				operation.SecondV.Content = append(
					operation.SecondV.Content, arithmetic.IntToString(token.Value[index]))
			}
		} else {
			operation.SecondV.Content = []string{token.Value}
		}
	}

	return 0
}

// processRange Process range by value processor principles.
// tokens Tokens to process.
func (i *Interpreter) processRange(tokens *vector.Vector) {
	/* Check parentheses range. */
	for true {
		_range, found := parser.DecomposeBrace(tokens, grammar.TokenLParenthes,
			grammar.TokenRParenthes, true)

		/* Parentheses are not found! */
		if found == -1 {
			break
		}

		val := i.processValue(_range)
		if val.Array {
			tokens.Insert(found, objects.Token{
				Value: grammar.TokenLBrace,
				Type:  fract.TypeBrace,
			})
			for index := range val.Content {
				found++
				tokens.Insert(found, objects.Token{
					Value: val.Content[index],
					Type:  fract.TypeValue,
				})
				found++
				tokens.Insert(found, objects.Token{
					Value: grammar.TokenComma,
					Type:  fract.TypeComma,
				})
			}
			found++
			tokens.Insert(found, objects.Token{
				Value: grammar.TokenRBrace,
				Type:  fract.TypeBrace,
			})
		} else {
			tokens.Insert(found, objects.Token{
				Value: val.Content[0],
				Type:  fract.TypeValue,
			})
		}
	}
}

// processArrayValue Process array value.
// tokens Tokens.
func (i *Interpreter) processArrayValue(tokens *vector.Vector) objects.Value {
	value := objects.Value{
		Array: true,
	}

	first := tokens.Vals[0].(objects.Token)

	// Initializer?
	if first.Value == grammar.TokenLBracket {
		valueList := tokens.Sublist(1, len(tokens.Vals)-2)

		if len(valueList.Vals) == 0 {
			fract.Error(first, "Size is not defined!")
		}

		value := i.processValue(valueList)
		if value.Array {
			fract.Error(first, "Arrays is not used in array constructors!")
		} else if arithmetic.IsFloatValue(value.Content[0]) {
			fract.Error(first, "Float values is not used in array constructors!")
		}

		val, _ := arithmetic.ToInt64(value.Content[0])
		if val < 0 {
			fract.Error(first, "Value is not lower than zero!")
		}
		value.Content = make([]string, val)
		for index := range value.Content {
			value.Content[index] = "0"
		}
		return value
	}

	comma := 1
	for index := 1; index < len(tokens.Vals)-1; index++ {
		current := tokens.Vals[index].(objects.Token)
		if current.Type == fract.TypeComma {
			lst := tokens.Sublist(comma, index-comma)
			if len(lst.Vals) == 0 {
				fract.Error(first, "Value is not defined!")
			}
			val := i.processValue(lst)
			value.Content = append(value.Content, val.Content...)
			if !value.Charray {
				value.Charray = val.Charray
			}
			comma = index + 1
		}
	}

	if comma < len(tokens.Vals)-1 {
		lst := tokens.Sublist(comma, len(tokens.Vals)-comma-1)
		if len(lst.Vals) == 0 {
			fract.Error(first, "Value is not defined!")
		}
		val := i.processValue(lst)
		value.Content = append(value.Content, val.Content...)
		if !value.Charray {
			value.Charray = val.Charray
		}
	}

	return value
}

// processValue Process value.
// tokens Tokens to process.
func (i *Interpreter) processValue(tokens *vector.Vector) objects.Value {
	value := objects.Value{
		Content: []string{"0"},
		Array:   false,
	}

	i.processRange(tokens)

	// Is conditional expression?
	if i.isConditional(tokens) {
		value.Content = []string{arithmetic.IntToString(i.processCondition(tokens))}
		return value
	}

	// Decompose arithmetic operations.
	priorityIndex := parser.IndexProcessPriority(tokens)
	looped := priorityIndex != -1
	for priorityIndex != -1 {
		var operation objects.ArithmeticProcess

		operation.First = tokens.Vals[priorityIndex-1].(objects.Token)
		priorityIndex -= i._processValue(true, &operation,
			tokens, priorityIndex-1)
		operation.Operator = tokens.Vals[priorityIndex].(objects.Token)

		operation.Second = tokens.Vals[priorityIndex+1].(objects.Token)
		priorityIndex -= i._processValue(false, &operation,
			tokens, priorityIndex+1)

		resultValue := arithmetic.SolveArithmeticProcess(operation)

		operation.Operator.Value = grammar.TokenPlus
		operation.Second = tokens.Vals[priorityIndex+1].(objects.Token)
		operation.FirstV = value
		operation.SecondV = resultValue

		resultValue = arithmetic.SolveArithmeticProcess(operation)
		value = resultValue

		// Remove processed processes.
		tokens.RemoveRange(priorityIndex-1, 3)
		tokens.Insert(priorityIndex-1, objects.Token{Value: "0"})

		// Find next operator.
		priorityIndex = parser.IndexProcessPriority(tokens)
	}

	// One value?
	if !looped {
		var operation objects.ArithmeticProcess
		operation.First = tokens.Vals[0].(objects.Token)
		operation.FirstV.Array = true // Ignore nil control if function call.
		i._processValue(true, &operation, tokens, 0)
		value = operation.FirstV
	}

	return value
}
