package obj

// Token instance.
type Token struct {
	File   *File
	Val    string
	Type   uint8
	Line   int
	Column int
}
