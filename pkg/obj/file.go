package obj

import "os"

// Source file instance.
type File struct {
	Path  string
	File  *os.File
	Lines []string
}
