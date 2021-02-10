/*
	processValue Function
*/

package interpreter

import (
	"strings"

	"../fract"
	"../fract/arithmetic"
	"../fract/dt"
	"../fract/name"
	"../grammar"
	"../objects"
	"../parser"
	"../utilities/vector"
)

// processValue Process value.
// tokens Tokens.
func (i *Interpreter) processValue(tokens *vector.Vector) objects.Value {
	// Is array expression?
	first := tokens.First().(objects.Token)
	if first.Type == fract.TypeBrace && (first.Value == grammar.TokenLBrace ||
		first.Value == grammar.TokenLBracket) {
		return i.processArrayValue(tokens)
	}

	/* Check parentheses range. */
	for true {
		_range, found := parser.DecomposeBrace(tokens, grammar.TokenLParenthes,
			grammar.TokenRParenthes)

		/* Parentheses are not found! */
		if found == -1 {
			break
		}

		var _token objects.Token
		_token.Value = i.processValue(&_range).Content[0]
		_token.Type = fract.TypeValue
		tokens.Insert(found, _token)
	}

	var value objects.Value
	value.Type = fract.VTInteger

	// Is conditional expression?
	if i.isConditional(tokens) {
		value.Content = []string{arithmetic.IntToString(i.processCondition(tokens))}
		return value
	}

	// Decompose arithmetic operations.
	operations := parser.DecomposeArithmeticProcesses(tokens)

	// Process arithmetic operation.
	priorityIndex := parser.IndexProcessPriority(&operations)
	for priorityIndex != -1 {
		var operation objects.ArithmeticProcess
		operation.First = operations.At(priorityIndex - 1).(objects.Token)
		operation.Operator = operations.At(priorityIndex).(objects.Token)
		operation.Second = operations.At(priorityIndex + 1).(objects.Token)

		// First value is a name?
		if operation.First.Type == fract.TypeName {
			index := name.VarIndexByName(i.vars, operation.First.Value)
			if index == -1 {
				fract.Error(operation.First,
					"Name is not defined!: "+operation.First.Value)
			}
			operation.First.Value = i.vars.At(index).(objects.Variable).Value[0]
		}

		// Second value is a name?
		if operation.Second.Type == fract.TypeName {
			index := name.VarIndexByName(i.vars, operation.Second.Value)
			if index == -1 {
				fract.Error(operation.Second,
					"Name is not defined!: "+operation.Second.Value)
			}
			operation.Second.Value = i.vars.At(index).(objects.Variable).Value[0]
		}

		_token := operations.At(priorityIndex - 1).(objects.Token)
		operations.RemoveRange(priorityIndex-1, 3)
		_type, result := arithmetic.SolveArithmeticProcess(operation)
		value.Type = _type
		_token.Value = arithmetic.TypeToString(_type, result)
		_token.Type = fract.TypeValue
		operations.Insert(priorityIndex-1, _token)

		// Find next operator.
		priorityIndex = parser.IndexProcessPriority(&operations)
	}

	// Set value.
	first = operations.First().(objects.Token)

	// First value is a name?
	if first.Type == fract.TypeName && tokens.Len() == 1 {
		index := name.VarIndexByName(i.vars, first.Value)
		if index == -1 {
			fract.Error(first,
				"Name is not defined!: "+first.Value)
		}
		variable := i.vars.At(index).(objects.Variable)
		// Is Array?
		if variable.Array {
			value.Content = variable.Value
			if dt.IsFloatType(variable.Type) {
				value.Type = fract.VTFloatArray
			} else {
				value.Type = fract.VTIntegerArray
			}
			return value
		}
		first.Value = variable.Value[0]
	}

	_value, err := arithmetic.ToFloat64(first.Value)
	if err != nil {
		fract.Error(first, "Value out of range!")
	}
	if arithmetic.IsFloatValue(first.Value) {
		value.Type = fract.VTFloat
	}
	value.Content = []string{arithmetic.TypeToString(value.Type, _value)}

	/* Set type to float if... */
	if value.Type != fract.VTFloat &&
		(strings.Index(value.Content[0], grammar.TokenDot) != -1 ||
			strings.Index(value.Content[0], grammar.TokenDot) != -1) {
		value.Type = fract.VTFloat
	}

	return value
}
