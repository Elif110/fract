package parser

import (
	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/grammar"
	obj "github.com/fract-lang/fract/pkg/objects"
)

// IndexProcessPriority Find index of priority operator.
// Returns index of operator if found, returns -1 if not.
//
// tokens Tokens to search.
func IndexProcessPriority(tokens []obj.Token) int {
	bracket := 0

	for index, token := range tokens {
		if token.Type == fract.TypeBrace {
			if token.Value == grammar.TokenLBracket ||
				token.Value == grammar.TokenLBrace ||
				token.Value == grammar.TokenLParenthes {
				bracket++
			} else {
				bracket--
			}
		}

		if bracket > 0 {
			continue
		}

		// Exponentiation or shifts.
		if token.Value == grammar.LeftBinaryShift || token.Value == grammar.RightBinaryShift ||
			token.Value == grammar.TokenCaret {
			return index
		}

		modulus := fract.TypeNone
		multiplyOrDivive := fract.TypeNone
		additionOrSubtraction := fract.TypeNone

		if token.Value == grammar.TokenPercent { // Modulus.
			if modulus == fract.TypeNone {
				modulus = index
			}
		} else if token.Value == grammar.TokenStar ||
			token.Value == grammar.TokenSlash ||
			token.Value == grammar.TokenBackslash ||
			token.Value == grammar.IntegerDivision ||
			token.Value == grammar.IntegerDivideWithBigger { // Multiply or division.
			if multiplyOrDivive == fract.TypeNone {
				multiplyOrDivive = index
			}
		} else if token.Value == grammar.TokenPlus ||
			token.Value == grammar.TokenMinus { // Addition or subtraction.
			if additionOrSubtraction == fract.TypeNone {
				additionOrSubtraction = index
			}
		}

		if modulus != fract.TypeNone {
			if modulus == len(tokens)-1 {
				fract.Error(tokens[modulus], "Operator defined, but for what?")
			}
			return modulus
		} else if multiplyOrDivive != fract.TypeNone {
			if multiplyOrDivive == len(tokens)-1 {
				fract.Error(tokens[multiplyOrDivive], "Operator defined, but for what?")
			}
			return multiplyOrDivive
		} else if additionOrSubtraction != fract.TypeNone {
			if additionOrSubtraction == len(tokens)-1 {
				fract.Error(tokens[additionOrSubtraction], "Operator defined, but for what?")
			}
			return additionOrSubtraction
		}
	}

	return fract.TypeNone
}
