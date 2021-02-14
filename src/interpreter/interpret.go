/*
	Interpret Function
*/

package interpreter

import "../utilities/vector"

// Interpret Interpret file.
func (i *Interpreter) Interpret() {
	// Lexer is finished.
	if i.lexer.Finished {
		return
	}

	/* Interpret all lines. */
	for !i.lexer.Finished {
		cacheTokens := i.lexer.Next()

		// cacheTokens are empty?
		if len(cacheTokens.Vals) == 0 {
			continue
		}

		i.tokens.Append(cacheTokens)
	}

	i.tokenLen = len(i.tokens.Vals)
	for ; i.index < i.tokenLen; i.index++ {
		i.processTokens(i.tokens.Vals[i.index].(*vector.Vector), true)
	}

	if i.blockCount > 0 { // Check blocks.
		i.lexer.Line--
		i.lexer.Error("Block is expected ending...")
	}
}
