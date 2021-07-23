package obj

import (
	"os"

	"github.com/fract-lang/fract/pkg/value"
)

// Var instance.
type Var struct {
	Name      string
	Ln        int // Line of define.
	V         value.Val
	Const     bool
	Protected bool
}

// Func instance.
type Func struct {
	Name          string
	Src           interface{}
	Ln            int // Line of define.
	Tks           []Tokens
	Params        []Param
	DefParamCount int
	Protected     bool
}

// Param instance.
type Param struct {
	Defval value.Val
	Name   string
	Params bool
}

// Source file instance.
type File struct {
	P   string
	F   *os.File
	Lns []string
}
