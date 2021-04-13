/*
	processImport Function.
*/

package interpreter

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/grammar"
	"github.com/fract-lang/fract/pkg/objects"
)

// processImport Process import.
// tokens Tokens to process.
func (i *Interpreter) processImport(tokens []objects.Token) {
	if len(tokens) == 1 {
		fract.Error(tokens[0], "Imported but what?")
	}

	if tokens[1].Type != fract.TypeName && (tokens[1].Type != fract.TypeValue ||
		(!strings.HasPrefix(tokens[1].Value, grammar.TokenDoubleQuote) &&
			!strings.HasPrefix(tokens[1].Value, grammar.TokenQuote))) {
		fract.Error(tokens[1], "Import path should be string or standard path!")
	}

	index := 1
	if len(tokens) > 2 {
		if tokens[1].Type == fract.TypeName {
			index = 2
		} else {
			fract.Error(tokens[1], "Alias is should be name!")
		}
	}

	if index == 1 && len(tokens) != 2 {
		fract.Error(tokens[2], "Invalid syntax!")
	} else if index == 2 && len(tokens) != 3 {
		fract.Error(tokens[3], "Invalid syntax!")
	}

	var path string
	if tokens[index].Type == fract.TypeName {
		if !strings.HasPrefix(tokens[index].Value, "std") {
			fract.Error(tokens[index], "Standard import should be starts with 'std' directory.")
		}
		path = strings.ReplaceAll(tokens[index].Value, grammar.TokenDot, string(os.PathSeparator))
	} else {
		path = tokens[0].File.Path[:strings.LastIndex(tokens[0].File.Path, string(os.PathSeparator))+1] +
			i.processValue(&[]objects.Token{tokens[index]}).Content[0].Data
	}

	info, err := os.Stat(path)

	// Exists directory?
	if err != nil || !info.IsDir() {
		fract.Error(tokens[index], "Directory not found/access!")
	}

	content, err := ioutil.ReadDir(path)
	if err != nil {
		fract.Error(tokens[1], "There is a problem on import: "+err.Error())
	}

	var name string
	if index == 1 {
		name = info.Name()
	} else {
		name = tokens[1].Value
	}

	for _, current := range content {
		// Skip directories.
		if current.IsDir() || !strings.HasSuffix(current.Name(), fract.FractExtension) {
			continue
		}

		New(path, path+string(os.PathSeparator)+current.Name()).Import(i, name)
	}
}