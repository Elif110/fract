package parser

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
)

// Import content into destination interpeter.
func (p *Parser) Import() {
	// Interpret all lines.
	for p.index = 1; p.index < len(p.Tokens); p.index++ {
		switch tokens := p.Tokens[p.index]; tokens[0].Type {
		case fract.Var:
			p.vardec(tokens)
		case fract.Fn:
			p.funcdec(tokens)
		case fract.Struct:
			p.structdec(tokens)
		case fract.Class:
			p.classdec(tokens)
		case fract.Import: // Import.
			src := new(Parser)
			src.AddBuiltInFuncs()
			src.processImport(tokens)
			p.defs.Vars = append(p.defs.Vars, src.defs.Vars...)
			p.defs.Funcs = append(p.defs.Funcs, src.defs.Funcs...)
			p.packages = append(p.packages, src.packages...)
		case fract.Macro: // Macro.
			p.processPragma(tokens)
			if p.loopCount != -1 { // Breaked import.
				return
			}
		}
	}
}

// Information of import.
type importInfo struct {
	name string  // Package name.
	src  *Parser // Source of package.
	line int     // Defined line.
}

func (p *Parser) processImport(tokens []obj.Token) {
	if len(tokens) == 1 {
		fract.IPanic(tokens[0], obj.SyntaxPanic, "Import path is not given!")
	}
	if tokens[1].Type != fract.Name && (tokens[1].Type != fract.Value || tokens[1].Val[0] != '"' && tokens[1].Val[0] != '.') {
		fract.IPanic(tokens[1], obj.ValuePanic, "Import path should be string or standard path!")
	}
	j := 1
	if len(tokens) > 2 {
		if tokens[1].Type == fract.Name {
			j = 2
		} else {
			fract.IPanic(tokens[1], obj.NamePanic, "Alias is should be a invalid name!")
		}
	}
	if j == 1 && len(tokens) != 2 {
		fract.IPanic(tokens[2], obj.SyntaxPanic, "Invalid syntax!")
	} else if j == 2 && len(tokens) != 3 {
		fract.IPanic(tokens[3], obj.SyntaxPanic, "Invalid syntax!")
	}
	src := &Parser{}
	src.AddBuiltInFuncs()
	var impPath string
	if tokens[j].Type == fract.Name {
		switch tokens[j].Val {
		default:
			impPath = strings.ReplaceAll(fract.StdLib+"/."+tokens[j].Val, ".", string(os.PathSeparator))
		}
	} else {
		impPath = tokens[0].File.Path[:strings.LastIndex(tokens[0].File.Path, string(os.PathSeparator))+1] + p.processValTokens([]obj.Token{tokens[j]}).String()
	}
	impPath = path.Join(fract.ExecutablePath, impPath)
	info, err := os.Stat(impPath)
	// Exists directory?
	if impPath != "" && (err != nil || !info.IsDir()) {
		fract.IPanic(tokens[j], obj.PlainPanic, "Directory not found/access!")
	}
	infos, err := ioutil.ReadDir(impPath)
	if err != nil {
		fract.IPanic(tokens[1], obj.PlainPanic, "There is a problem on import: "+err.Error())
	}
	for _, i := range infos {
		// Skip directories.
		if i.IsDir() || !strings.HasSuffix(i.Name(), fract.Extension) {
			continue
		}
		impSrc := New(impPath + string(os.PathSeparator) + i.Name())
		impSrc.loopCount = -1 //! Tag as import source.
		impSrc.ready()
		impSrc.AddBuiltInFuncs()
		builtinFuncLen := len(impSrc.defs.Funcs)
		impSrc.Import()
		impSrc.importPackage() // Import other package files.
		impSrc.loopCount = 0
		src.defs.Funcs = append(src.defs.Funcs, impSrc.defs.Funcs[builtinFuncLen:]...)
		src.defs.Vars = append(src.defs.Vars, impSrc.defs.Vars...)
		src.packages = append(src.packages, impSrc.packages...)
		src.packageName = impSrc.packageName
		break
	}
	p.packages = append(p.packages, importInfo{name: src.packageName, src: src, line: tokens[0].Line})
}
