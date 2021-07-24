package fract

const (
	Ver         = "0.0.1"  // Version of Fract.
	Ext         = ".fract" // File extension of Fract.
	FloatFormat = "%g"     // Float format.

	None                uint8 = 0
	Ignore              uint8 = 1
	Comment             uint8 = 10
	Operator            uint8 = 11
	Value               uint8 = 12
	Brace               uint8 = 13
	Var                 uint8 = 14
	Name                uint8 = 15
	Comma               uint8 = 16
	If                  uint8 = 17
	Else                uint8 = 18
	StatementTerminator uint8 = 19
	Loop                uint8 = 20
	In                  uint8 = 21
	Break               uint8 = 22
	Continue            uint8 = 23
	Func                uint8 = 24
	Ret                 uint8 = 25
	Try                 uint8 = 26
	Catch               uint8 = 27
	Import              uint8 = 28
	Params              uint8 = 29
	Macro               uint8 = 30
	Defer               uint8 = 31
	Go                  uint8 = 32
	Colon               uint8 = 33
	Package             uint8 = 34
	Dot                 uint8 = 35

	LOOPBreak    uint8 = 1
	LOOPContinue uint8 = 2
	FUNCReturn   uint8 = 3
)

var (
	TryCount      int // Try-Catch count.
	ExecPath      string
	InteractiveSh bool // Interactive shell mode.
)
