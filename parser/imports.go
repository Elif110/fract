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
	for p.i = 1; p.i < len(p.Tks); p.i++ {
		switch tks := p.Tks[p.i]; tks[0].T {
		case fract.Var:
			p.vardec(tks)
		case fract.Fn:
			p.funcdec(tks)
		case fract.Struct:
			p.structdec(tks)
		case fract.Class:
			p.classdec(tks)
		case fract.Import: // Import.
			src := new(Parser)
			src.AddBuiltInFuncs()
			src.procImport(tks)
			p.defs.Vars = append(p.defs.Vars, src.defs.Vars...)
			p.defs.Funcs = append(p.defs.Funcs, src.defs.Funcs...)
			p.packages = append(p.packages, src.packages...)
		case fract.Macro: // Macro.
			p.procPragma(tks)
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
	ln   int     // Defined line.
}

func (p *Parser) procImport(tks []obj.Token) {
	if len(tks) == 1 {
		fract.IPanic(tks[0], obj.SyntaxPanic, "Import path is not given!")
	}
	if tks[1].T != fract.Name && (tks[1].T != fract.Value || tks[1].V[0] != '"' && tks[1].V[0] != '.') {
		fract.IPanic(tks[1], obj.ValuePanic, "Import path should be string or standard path!")
	}
	j := 1
	if len(tks) > 2 {
		if tks[1].T == fract.Name {
			j = 2
		} else {
			fract.IPanic(tks[1], obj.NamePanic, "Alias is should be a invalid name!")
		}
	}
	if j == 1 && len(tks) != 2 {
		fract.IPanic(tks[2], obj.SyntaxPanic, "Invalid syntax!")
	} else if j == 2 && len(tks) != 3 {
		fract.IPanic(tks[3], obj.SyntaxPanic, "Invalid syntax!")
	}
	src := &Parser{}
	src.AddBuiltInFuncs()
	var imppath string
	if tks[j].T == fract.Name {
		switch tks[j].V {
		default:
			imppath = strings.ReplaceAll(fract.StdLib+"/."+tks[j].V, ".", string(os.PathSeparator))
		}
	} else {
		imppath = tks[0].F.P[:strings.LastIndex(tks[0].F.P, string(os.PathSeparator))+1] + p.procValTks([]obj.Token{tks[j]}).String()
	}
	imppath = path.Join(fract.ExecPath, imppath)
	info, err := os.Stat(imppath)
	// Exists directory?
	if imppath != "" && (err != nil || !info.IsDir()) {
		fract.IPanic(tks[j], obj.PlainPanic, "Directory not found/access!")
	}
	infos, err := ioutil.ReadDir(imppath)
	if err != nil {
		fract.IPanic(tks[1], obj.PlainPanic, "There is a problem on import: "+err.Error())
	}
	for _, i := range infos {
		// Skip directories.
		if i.IsDir() || !strings.HasSuffix(i.Name(), fract.Extension) {
			continue
		}
		isrc := New(imppath + string(os.PathSeparator) + i.Name())
		isrc.loopCount = -1 //! Tag as import source.
		isrc.ready()
		isrc.AddBuiltInFuncs()
		bifl := len(isrc.defs.Funcs)
		isrc.Import()
		isrc.importPackage() // Import other package files.
		isrc.loopCount = 0
		src.defs.Funcs = append(src.defs.Funcs, isrc.defs.Funcs[bifl:]...)
		src.defs.Vars = append(src.defs.Vars, isrc.defs.Vars...)
		src.packages = append(src.packages, isrc.packages...)
		src.pkg = isrc.pkg
		break
	}
	p.packages = append(p.packages, importInfo{name: src.pkg, src: src, ln: tks[0].Ln})
}
