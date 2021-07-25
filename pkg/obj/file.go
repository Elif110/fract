package obj

import "os"

// Source file instance.
type File struct {
	P   string
	F   *os.File
	Lns []string
}
