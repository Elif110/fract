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

func (p Panic) Panic(exit bool) {
	if exit {
		fmt.Println(p)
		os.Exit(1)
	}
	panic(p)
}
