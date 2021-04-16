/*
	Import Function.
*/

package interpreter

import (
	"unicode"

	"github.com/fract-lang/fract/pkg/fract"
)

// Import content into destination interpeter.
func (i *Interpreter) Import() {
	i.ready()

	varLen := len(i.variables)
	funcLen := len(i.functions)

	checkFunction := func() {
		funcLen++
		if !unicode.IsUpper(rune(i.functions[funcLen-1].Name[0])) {
			funcLen--
			i.functions = i.functions[:funcLen]
		}
	}

	checkVariable := func() {
		varLen++
		if !unicode.IsUpper(rune(i.variables[varLen-1].Name[0])) {
			varLen--
			i.variables = i.variables[:varLen]
		}
	}

	// Interpret all lines.
	for i.index = 0; i.index < len(i.Tokens); i.index++ {
		tokens := i.Tokens[i.index]
		switch first := tokens[0]; first.Type {
		case fract.TypeProtected: // Protected declaration.
			if len(tokens) < 2 {
				fract.Error(first, "Protected but what is it protected?")
			}
			second := tokens[1]
			tokens = tokens[1:]
			if second.Type == fract.TypeVariable { // Variable definition.
				i.processVariableDefinition(tokens, true)
				checkVariable()
			} else if second.Type == fract.TypeFunction { // Function definition.
				i.processFunction(tokens, true)
				checkFunction()
			} else {
				fract.Error(second, "Syntax error, you can protect only deletable objects!")
			}
		case fract.TypeVariable: // Variable definition.
			i.processVariableDefinition(tokens, false)
			checkVariable()
		case fract.TypeFunction: // Function definiton.
			i.processFunction(tokens, false)
			checkFunction()
		case fract.TypeImport: // Import.
			source := Interpreter{}
			source.processImport(tokens)

			i.variables = append(i.variables, source.variables...)
			i.functions = append(i.functions, source.functions...)
			i.Imports = append(i.Imports, source.Imports...)
		}
	}
}
