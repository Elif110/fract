package fract

import (
	"os"
	"fmt"
	"strings"

	"github.com/fract-lang/fract/pkg/obj"
	"github.com/fract-lang/fract/pkg/str"
)

func PanicC(f *obj.File, col, ln int, t, m string) {
	e := obj.Panic{
		Msg: fmt.Sprintf("File: %s\nPosition: %d:%d\n    %s\n%s^\n%s: %s",
			f.Path, ln, col, strings.ReplaceAll(f.Lines[ln-1], "\t", " "),
			str.Full(4+col-2, ' '), t, m),
		Type: t,
	}
	if TryCount > 0 {
		panic(e)
	}
	e.Panic(!InteractiveShell)
}

func Panic(tk obj.Token, t, m string) { PanicC(tk.File, tk.Column, tk.Line, t, m) }

// Interpreter panic.
func IPanicC(f *obj.File, ln, col int, t, m string) {
	e := obj.Panic{
		Msg: fmt.Sprintf("File: %s\nPosition: %d:%d\n    %s\n%s^\n%s: %s",
			f.Path, ln, col, strings.ReplaceAll(f.Lines[ln-1], "\t", " "),
			str.Full(4+col-2, ' '), t, m),
		Type: t,
	}
	e.Panic(!InteractiveShell)
}

// Interpreter panic.
func IPanic(tk obj.Token, t, m string) { IPanicC(tk.File, tk.Line, tk.Column, t, m) }

// Error is text interpreter panic.
func Error(f *obj.File, ln, col int, m string) {
	fmt.Printf("File: %s\nPosition: %d:%d\n%s\n", f.Path, ln, col, m)
	os.Exit(1)
}
