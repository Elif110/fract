package str

import (
	"strconv"
	"strings"
)

// Full returns r as string by length.
func Full(len int, r rune) string {
	var sb strings.Builder
	for len >= 0 {
		sb.WriteRune(r)
		len--
	}
	return sb.String()
}

// Parse string to arithmetic oop.
func Conv(v string) float64 {
	switch v {
	case "true":
		return 1
	case "false":
		return 0
	default:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}
}
