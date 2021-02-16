/*
	ReadyLines Function.
*/

package interpreter

import (
	"strings"

	"../objects"
	"../utilities/vector"
)

// ReadyLines Ready lines to process.
// lines Lines to ready.
func ReadyLines(lines []string) *vector.Vector {
	readyLines := vector.New()
	for index := 0; index < len(lines); index++ {
		readyLines.Vals = append(readyLines.Vals, objects.CodeLine{Line: index + 1,
			Text: strings.TrimRight(lines[index], " \t\n\r")})
	}
	return readyLines
}