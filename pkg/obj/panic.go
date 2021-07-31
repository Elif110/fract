package obj

import (
	"fmt"
	"os"
)

const (
	PlainPanic        = "Panic"
	NamePanic         = "NamePanic"
	MemoryPanic       = "MemoryPanic"
	SyntaxPanic       = "SyntaxPanic"
	ValuePanic        = "ValuePanic"
	OutOfRangePanic   = "OutOfRangePanic"
	ArithmeticPanic   = "ArithmeticPanic"
	DivideByZeroPanic = "DivideByZeroPanic"
)

type Panic struct {
	Msg  string
	Type string
}

func (p Panic) String() string { return p.Msg }

func (p Panic) Panic() { fmt.Println(p); os.Exit(1) }
